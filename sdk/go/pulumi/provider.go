// Copyright 2016-2021, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pulumi

import (
	"context"
	"flag"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v2/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/logging"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/rpcutil"

	pbempty "github.com/golang/protobuf/ptypes/empty"
	pulumirpc "github.com/pulumi/pulumi/sdk/v2/proto/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ProviderMain(provider ProviderArgs) error {
	if provider.Name == "" {
		return errors.New("provider.Name must not be empty")
	}
	if provider.Version == "" {
		return errors.New("provider.Version must not be empty")
	}

	var tracing string
	flag.StringVar(&tracing, "tracing", "", "Emit tracing to a Zipkin-compatible tracing endpoint")
	flag.Parse()

	// Initialize loggers before going any further.
	logging.InitLogging(false, 0, false)
	cmdutil.InitTracing(provider.Name, provider.Name, tracing)

	// Read the non-flags args and connect to the engine.
	args := flag.Args()
	if len(args) == 0 {
		return errors.New("fatal: could not connect to host RPC; missing argument")
	}

	prov := &providerServer{
		provider:   provider,
		engineAddr: args[0],
	}

	// Fire up a gRPC server, letting the kernel choose a free port for us.
	port, done, err := rpcutil.Serve(0, nil, []func(*grpc.Server) error{
		func(srv *grpc.Server) error {
			pulumirpc.RegisterResourceProviderServer(srv, prov)
			return nil
		},
	}, nil)
	if err != nil {
		return errors.Errorf("fatal: %v", err)
	}

	// The resource provider protocol requires that we now write out the port we have chosen to listen on.
	fmt.Printf("%d\n", port)

	// Finally, wait for the server to stop serving.
	if err := <-done; err != nil {
		return errors.Errorf("fatal: %v", err)
	}

	return nil
}

type providerServer struct {
	provider   ProviderArgs
	engineAddr string
}

type ProviderArgs struct {
	Name    string
	Version string
	Schema  []byte

	ConstructF func(ctx *Context, typ, name string, inputs *ConstructInputs,
		options ResourceOption) (ConstructResult, error)

	// TODO add all the other gRPC methods.
}

type constructInput struct {
	value  interface{}
	secret bool
	deps   []Resource
}

type ConstructInputs struct {
	inputs map[string]constructInput
}

func (inputs *ConstructInputs) Map() Map {
	result := Map{}
	for k, v := range inputs.inputs {
		output := newOutput(anyOutputType, v.deps...)
		output.resolve(v.value, true /*known*/, v.secret, nil)
		result[k] = output
	}
	return result
}

func (inputs *ConstructInputs) SetArgs(args interface{}) error {
	if args == nil {
		return errors.New("args must not be nil")
	}
	argsV := reflect.ValueOf(args)
	typ := argsV.Type()
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return errors.New("args must be a pointer to a struct")
	}
	argsV, typ = argsV.Elem(), typ.Elem()

	for k, v := range inputs.inputs {
		for i := 0; i < typ.NumField(); i++ {
			fieldV := argsV.Field(i)
			if !fieldV.CanSet() {
				continue
			}
			field := typ.Field(i)
			tag, has := field.Tag.Lookup("pulumi")
			if !has || tag != k {
				continue
			}

			toOutputMethodName := "To" + strings.TrimSuffix(field.Type.Name(), "Input") + "OutputWithContext"
			toOutputMethod, found := field.Type.MethodByName(toOutputMethodName)
			if !found {
				continue
			}
			mt := toOutputMethod.Type
			if mt.NumIn() != 1 || mt.In(0) != contextType || mt.NumOut() != 1 {
				continue
			}
			outputType := mt.Out(0)
			if !outputType.Implements(reflect.TypeOf((*Output)(nil)).Elem()) {
				continue
			}

			output := newOutput(outputType, v.deps...)
			output.resolve(v.value, true /*known*/, v.secret, nil)
			fieldV.Set(reflect.ValueOf(output))
		}
	}

	return nil
}

type ConstructResult struct {
	URN   URNInput
	State Input
}

// Construct creates a new instance of the provided component resource and returns its state.
func (p *providerServer) Construct(ctx context.Context,
	req *pulumirpc.ConstructRequest) (*pulumirpc.ConstructResponse, error) {

	if p.provider.ConstructF == nil {
		return nil, errors.Errorf("unknown resource type %s", req.GetType())
	}

	// Configure the RunInfo.
	runInfo := RunInfo{
		Project:     req.GetProject(),
		Stack:       req.GetStack(),
		Config:      req.GetConfig(),
		Parallel:    int(req.GetParallel()),
		DryRun:      req.GetDryRun(),
		MonitorAddr: req.GetMonitorEndpoint(),
		EngineAddr:  p.engineAddr,
		Mocks:       nil,
	}
	pulumiCtx, err := NewContext(ctx, runInfo)
	if err != nil {
		return nil, errors.Wrap(err, "constructing run context")
	}

	// Deserialize the inputs and apply appropriate dependencies.
	inputs := &ConstructInputs{inputs: map[string]constructInput{}}
	inputDependencies := req.GetInputDependencies()
	deserializedInputs, err := plugin.UnmarshalProperties(
		req.GetInputs(),
		plugin.MarshalOptions{KeepSecrets: true, KeepResources: true, KeepUnknowns: req.GetDryRun()},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling inputs")
	}
	for key, input := range deserializedInputs {
		k := string(key)
		var deps []Resource
		if inputDeps, ok := inputDependencies[k]; ok {
			deps = make([]Resource, len(inputDeps.GetUrns()))
			for i, depURN := range inputDeps.GetUrns() {
				deps[i] = newDependencyResource(URN(depURN))
			}
		}

		val, secret, err := unmarshalPropertyValue(pulumiCtx, input)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshaling input %s", k)
		}

		inputs.inputs[k] = constructInput{
			value:  val,
			secret: secret,
			deps:   deps,
		}
	}

	// Rebuild the resource options.
	aliases := make([]Alias, len(req.GetAliases()))
	for i, urn := range req.GetAliases() {
		aliases[i] = Alias{URN: URN(urn)}
	}
	dependencies := make([]Resource, len(req.GetDependencies()))
	for i, urn := range req.GetDependencies() {
		dependencies[i] = newDependencyResource(URN(urn))
	}
	providers := make(map[string]ProviderResource, len(req.GetProviders()))
	for pkg, ref := range req.GetProviders() {
		// Parse the URN and ID out of the provider reference.
		lastSep := strings.LastIndex(ref, "::")
		if lastSep == -1 {
			return nil, errors.Errorf("expected '::' in provider reference %s", ref)
		}
		urn := ref[0:lastSep]
		id := ref[lastSep+2:]
		providers[pkg] = newDependencyProviderResource(URN(urn), ID(id))
	}
	var parent Resource
	if req.GetParent() != "" {
		parent = newDependencyResource(URN(req.GetParent()))
	}
	opts := resourceOption(func(ro *resourceOptions) {
		ro.Aliases = aliases
		ro.DependsOn = dependencies
		ro.Protect = req.GetProtect()
		ro.Providers = providers
		ro.Parent = parent
	})

	result, err := p.provider.ConstructF(pulumiCtx, req.GetType(), req.GetName(), inputs, opts)
	if err != nil {
		return nil, err
	}

	// Ensure all outstanding RPCs have completed before proceeding. Also, prevent any new RPCs from happening.
	pulumiCtx.waitForRPCs()
	if pulumiCtx.rpcError != nil {
		return nil, errors.Wrap(pulumiCtx.rpcError, "waiting for RPCs")
	}

	rpcURN, _, _, err := result.URN.ToURNOutput().awaitURN(ctx)
	if err != nil {
		return nil, err
	}

	// Serialize all state properties, first by awaiting them, and then marshaling them to the requisite gRPC values.
	resolvedProps, propertyDeps, _, err := marshalInputs(result.State)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling properties")
	}

	// Marshal all properties for the RPC call.
	keepUnknowns := req.GetDryRun()
	rpcProps, err := plugin.MarshalProperties(
		resolvedProps,
		plugin.MarshalOptions{KeepSecrets: true, KeepUnknowns: keepUnknowns, KeepResources: pulumiCtx.keepResources})
	if err != nil {
		return nil, errors.Wrap(err, "marshaling properties")
	}

	// Convert the property dependencies map for RPC and remove duplicates.
	rpcPropertyDeps := make(map[string]*pulumirpc.ConstructResponse_PropertyDependencies)
	for k, deps := range propertyDeps {
		sort.Slice(deps, func(i, j int) bool { return deps[i] < deps[j] })

		urns := make([]string, 0, len(deps))
		for i, d := range deps {
			if i > 0 && urns[i-1] == string(d) {
				continue
			}
			urns = append(urns, string(d))
		}

		rpcPropertyDeps[k] = &pulumirpc.ConstructResponse_PropertyDependencies{
			Urns: urns,
		}
	}

	return &pulumirpc.ConstructResponse{
		Urn:               string(rpcURN),
		State:             rpcProps,
		StateDependencies: rpcPropertyDeps,
	}, nil
}

// GetSchema returns the JSON-encoded schema for this provider's package.
func (p *providerServer) GetSchema(ctx context.Context,
	req *pulumirpc.GetSchemaRequest) (*pulumirpc.GetSchemaResponse, error) {
	if v := req.GetVersion(); v != 0 {
		return nil, errors.Errorf("unsupported schema version %d", v)
	}
	return &pulumirpc.GetSchemaResponse{Schema: string(p.provider.Schema)}, nil
}

// CheckConfig validates the configuration for this provider.
func (p *providerServer) CheckConfig(ctx context.Context,
	req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return nil, status.Error(codes.Unimplemented, "CheckConfig is not yet implemented")
}

// DiffConfig diffs the configuration for this provider.
func (p *providerServer) DiffConfig(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return nil, status.Error(codes.Unimplemented, "DiffConfig is not yet implemented")
}

// Configure configures the resource provider with "globals" that control its behavior.
func (p *providerServer) Configure(ctx context.Context,
	req *pulumirpc.ConfigureRequest) (*pulumirpc.ConfigureResponse, error) {
	return &pulumirpc.ConfigureResponse{
		AcceptSecrets:   true,
		SupportsPreview: true,
		AcceptResources: true,
	}, nil
}

// Invoke dynamically executes a built-in function in the provider.
func (p *providerServer) Invoke(ctx context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	tok := req.GetTok()
	return nil, errors.Errorf("unknown Invoke token %q", tok)
}

// StreamInvoke dynamically executes a built-in function in the provider. The result is streamed
// back as a series of messages.
func (p *providerServer) StreamInvoke(req *pulumirpc.InvokeRequest,
	server pulumirpc.ResourceProvider_StreamInvokeServer) error {
	tok := req.GetTok()
	return errors.Errorf("unknown StreamInvoke token %q", tok)
}

// Check validates that the given property bag is valid for a resource of the given type and returns
// the inputs that should be passed to successive calls to Diff, Create, or Update for this
// resource. As a rule, the provider inputs returned by a call to Check should preserve the original
// representation of the properties as present in the program inputs. Though this rule is not
// required for correctness, violations thereof can negatively impact the end-user experience, as
// the provider inputs are using for detecting and rendering diffs.
func (p *providerServer) Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Check is not yet implemented")
}

// Diff checks what impacts a hypothetical update will have on the resource's properties.
func (p *providerServer) Diff(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Diff is not yet implemented")
}

// Create allocates a new instance of the provided resource and returns its unique ID afterwards.
// (The input ID must be blank.)  If this call fails, the resource must not have been created (i.e.,
// it is "transactional").
func (p *providerServer) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Create is not yet implemented")
}

// Read the current live state associated with a resource.  Enough state must be include in the
// inputs to uniquely identify the resource; this is typically just the resource ID, but may also
// include some properties.
func (p *providerServer) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Read is not yet implemented")
}

// Update updates an existing resource with new values.
func (p *providerServer) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Update is not yet implemented")
}

// Delete tears down an existing resource with the given ID.  If it fails, the resource is assumed
// to still exist.
func (p *providerServer) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return &pbempty.Empty{}, nil
}

// GetPluginInfo returns generic information about this plugin, like its version.
func (p *providerServer) GetPluginInfo(context.Context, *pbempty.Empty) (*pulumirpc.PluginInfo, error) {
	return &pulumirpc.PluginInfo{
		Version: p.provider.Version,
	}, nil
}

// Cancel signals the provider to gracefully shut down and abort any ongoing resource operations.
// Operations aborted in this way will return an error (e.g., `Update` and `Create` will either a
// creation error or an initialization error). Since Cancel is advisory and non-blocking, it is up
// to the host to decide how long to wait after Cancel is called before (e.g.)
// hard-closing any gRPC connection.
func (p *providerServer) Cancel(context.Context, *pbempty.Empty) (*pbempty.Empty, error) {
	return &pbempty.Empty{}, nil
}
