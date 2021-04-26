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

package provider

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"unicode"

	"github.com/pkg/errors"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

var contextType = reflect.TypeOf((*pulumi.Context)(nil))
var errorType = reflect.TypeOf((*error)(nil)).Elem()
var resourceType = reflect.TypeOf((*pulumi.Resource)(nil)).Elem()
var componentResourceType = reflect.TypeOf((*pulumi.ComponentResource)(nil)).Elem()

type componentInfo struct {
	factory      reflect.Value
	argsType     reflect.Type
	resourceType reflect.Type
	inputs       map[string]string
	outputs      map[string]string
}

// ComponentMain is an entrypoint for a resource provider plugin that implements `Construct` for component resources.
// Using it isn't required but can cut down significantly on the amount of boilerplate necessary to fire up a new
// resource provider for components.
func ComponentMainAuto(name, version string, componentFactories ...interface{}) error {
	// Maps component token to its construct function.
	components := make(map[string]componentInfo)

	for i, factory := range componentFactories {
		fv := reflect.ValueOf(factory)
		if fv.Kind() != reflect.Func {
			return errors.Errorf("componentFactories[%v] not a function", i)
		}

		ft := fv.Type()

		// TODO handle these cases:
		// ft.NumIn() == 2: ctx, name
		// ft.NumIn() == 3: ctx, name, ...options or args
		// ft.NumIn() == 4: ctx, name, args, ...options
		if ft.NumIn() != 4 {
			panic(errors.New("expected 4 inputs"))
		}
		if !ft.In(0).AssignableTo(contextType) {
			panic(errors.Errorf("first argument must be %v", contextType))
		}
		if ft.In(1).Kind() != reflect.String {
			panic(errors.New("second argument must be a string"))
		}

		argsType := ft.In(2)
		if argsType.Kind() == reflect.Ptr {
			argsType = argsType.Elem()
		}

		if ft.NumOut() != 2 {
			panic(errors.New("expected 2 return values"))
		}
		if !ft.Out(0).AssignableTo(componentResourceType) {
			panic(errors.New("first return type must be assignable to pulumi.ComponentResource"))
		}
		if !ft.Out(1).AssignableTo(errorType) {
			panic(errors.New("second return type must be assignable to error"))
		}

		resourceType := ft.Out(0)
		if resourceType.Kind() == reflect.Ptr {
			resourceType = resourceType.Elem()
		}

		// TODO error if the pkg part of the token doesn't match the specified package name?
		token, inputs, outputs, err := tokenAndDescriptions(fv, argsType.Name(), resourceType.Name())
		if err != nil {
			return err
		}

		components[token] = componentInfo{
			factory:      fv,
			argsType:     argsType,
			resourceType: resourceType,
			inputs:       inputs,
			outputs:      outputs,
		}
	}

	pkgSpec := genSchema(name, components)
	schemaJSON, err := json.MarshalIndent(pkgSpec, "", "    ")
	if err != nil {
		return err
	}

	var emitSchema bool
	for _, arg := range os.Args[1:] {
		if arg == "-schema" {
			emitSchema = true
			break
		}
	}
	if emitSchema {
		fmt.Println(string(schemaJSON))
		return nil
	}

	construct := func(ctx *pulumi.Context, typ, name string, inputs provider.ConstructInputs,
		options pulumi.ResourceOption) (*provider.ConstructResult, error) {
		info, ok := components[typ]
		if !ok {
			return nil, errors.Errorf("unknown resource type %s", typ)
		}

		// Copy the raw inputs to args. `inputs.CopyTo` uses the types and `pulumi:` tags
		// on the struct's fields to convert the raw values to the appropriate Input types.
		args := reflect.New(info.argsType)
		if err := inputs.CopyTo(args.Interface()); err != nil {
			return nil, err
		}

		// Create the component resource.
		in := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(name), args, reflect.ValueOf(options)}
		out := info.factory.Call(in)
		contract.Assertf(len(out) == 2, "expected two results")
		if err, ok := out[1].Interface().(error); ok && err != nil {
			return nil, err
		}

		// Return the component resource's URN and state. `NewConstructResult` automatically sets the
		// ConstructResult's state based on resource struct fields tagged with `pulumi:` tags with a value
		// that is convertible to `pulumi.Input`.
		return provider.NewConstructResult(out[0].Interface().(pulumi.ComponentResource))
	}

	return main(name, func(host *HostClient) (pulumirpc.ResourceProviderServer, error) {
		return &componentProvider{
			host:      host,
			name:      name,
			version:   version,
			schema:    schemaJSON,
			construct: construct,
		}, nil
	}, true /*schema*/)
}

func genSchema(name string, components map[string]componentInfo) schema.PackageSpec {
	// TODO provide a way to specify more metadata.
	pkg := schema.PackageSpec{
		Name: name,
		// TODO be smarter about these based on the types used, and/or provide a way for users to override.
		Language: map[string]json.RawMessage{
			"csharp": rawMessage(map[string]interface{}{
				"packageReferences": map[string]string{
					"Pulumi":     "3.*",
					"Pulumi.Aws": "4.*",
				},
			}),
			"nodejs": rawMessage(map[string]interface{}{
				"dependencies": map[string]string{
					"@pulumi/aws": "^4.0.0",
				},
				"devDependencies": map[string]string{
					"typescript": "^3.7.0",
				},
			}),
			"python": rawMessage(map[string]interface{}{
				"requires": map[string]string{
					"pulumi":     ">=3.0.0,<4.0.0",
					"pulumi-aws": ">=4.0.0,<5.0.0",
				},
				"readme": "",
			}),
			"go": rawMessage(map[string]interface{}{
				"generateResourceContainerTypes": true,
			}),
		},
	}

	// Add components.
	if len(components) > 0 {
		pkg.Resources = make(map[string]schema.ResourceSpec)
	}
	for token, component := range components {
		pkg.Resources[token] = genResourceSpec(component)
	}
	return pkg
}

func genResourceSpec(component componentInfo) schema.ResourceSpec {
	spec := schema.ResourceSpec{
		IsComponent: true,
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Description: component.outputs[""],
		},
	}

	// Inputs
	var requiredInputs []string
	inputs := make(map[string]schema.PropertySpec)
	for i := 0; i < component.argsType.NumField(); i++ {
		field := component.argsType.Field(i)
		tag := field.Tag.Get("pulumi")
		if tag == "" {
			continue
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		} else {
			requiredInputs = append(requiredInputs, tag)
		}

		inputs[tag] = schema.PropertySpec{
			TypeSpec:    genTypeSpec(fieldType, true /*input*/),
			Description: component.inputs[field.Name],
		}
	}
	spec.InputProperties = inputs
	spec.RequiredInputs = requiredInputs

	// Outputs
	var required []string
	outputs := make(map[string]schema.PropertySpec)
	for i := 0; i < component.resourceType.NumField(); i++ {
		field := component.resourceType.Field(i)
		tag := field.Tag.Get("pulumi")
		if tag == "" {
			continue
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			// If it's a resource output, it'll be a pointer, but we'll consider it to be always populated.
			if fieldType.AssignableTo(resourceType) {
				required = append(required, tag)
			}
			fieldType = fieldType.Elem()
		} else {
			required = append(required, tag)
		}

		outputs[tag] = schema.PropertySpec{
			TypeSpec:    genTypeSpec(fieldType, false /*input*/),
			Description: component.outputs[field.Name],
		}
	}
	spec.ObjectTypeSpec.Properties = outputs
	spec.ObjectTypeSpec.Required = required

	return spec
}

func genTypeSpec(fieldType reflect.Type, input bool) schema.TypeSpec {
	// github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3
	// TODO figure out a way to determine this for non-pulumi packages and atypical import paths.
	if reflect.PtrTo(fieldType).AssignableTo(resourceType) {
		components := strings.Split(fieldType.PkgPath(), "/")
		if len(components) >= 3 &&
			components[0] == "github.com" &&
			components[1] == "pulumi" &&
			strings.HasPrefix(components[2], "pulumi-") {
			pkg := strings.TrimPrefix(components[2], "pulumi-")
			if len(components) >= 4 && components[3] == "sdk" {
				components = components[4:]
				version := "v1"
				if strings.HasPrefix(components[0], "v") {
					version = components[0]
					components = components[1:]
				}
				if len(components) > 0 && components[0] == "go" {
					components = components[1:]
				}
				if len(components) > 0 && components[0] == pkg {
					components = components[1:]
				}

				// Turn into: /aws/v4.0.0/schema.json#/resources/aws:s3%2Fbucket:Bucket
				ref := fmt.Sprintf("/%s/%s.0.0/schema.json#/resources/%s:%s%s:%s", pkg, version, pkg, strings.Join(components, "%2F"), "%2F"+camel(fieldType.Name()), fieldType.Name())
				return schema.TypeSpec{Ref: ref}
			}

		}
	}

	// TODO other types
	switch fieldType.Name() {
	// number?
	case "int", "IntInput", "IntOutput":
		return schema.TypeSpec{Type: "integer"}
	case "bool", "BoolInput", "BoolOutput":
		return schema.TypeSpec{Type: "boolean"}
	case "string", "StringInput", "StringOutput":
		fallthrough
	default:
		return schema.TypeSpec{Type: "string"}
	}
}

// tokenAndDescriptions gets the token from the pulumi comment directive on the resoure type struct, and
// comments for the args and resource types.
func tokenAndDescriptions(fv reflect.Value, argsTypeName, resourceTypeName string) (string, map[string]string,
	map[string]string, error) {
	contract.Assertf(fv.Kind() == reflect.Func, "fv not a function")

	// Determine the source file of the function. We'll look for the args and resource structs
	// in the same file.
	filename, _ := runtime.FuncForPC(fv.Pointer()).FileLine(0)
	fset := token.NewFileSet()

	// Parse the file.
	parsed, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return "", nil, nil, err
	}

	pkg := &ast.Package{
		Name:  "Any",
		Files: map[string]*ast.File{filename: parsed},
	}

	// TODO don't bother using `doc` -- just look at the AST directly.

	importPath := filepath.Base(filename)
	funcDoc := doc.New(pkg, importPath, doc.AllDecls|doc.PreserveAST)

	var token string
	inputs, outputs := make(map[string]string), make(map[string]string)
	var sawInputs, sawOutputs bool
	for _, typ := range funcDoc.Types {
		switch typ.Name {
		case argsTypeName:
			getDescriptions(typ, inputs)
			sawInputs = true
		case resourceTypeName:
			token, err = getTokenFromPulumiDirective(typ)
			if err != nil {
				return "", nil, nil, err
			}
			getDescriptions(typ, outputs)
			sawOutputs = true
		default:
			if sawInputs && sawOutputs {
				break
			}
		}
	}

	// TODO provide some kind of warning if couldn't find args/resource types?

	return token, inputs, outputs, nil
}

func getDescriptions(typ *doc.Type, descriptions map[string]string) {
	if typ.Doc != "" {
		descriptions[""] = strings.TrimSpace(typ.Doc)
	}
	if len(typ.Decl.Specs) != 1 {
		return
	}
	typeSpec, ok := typ.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return
	}
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return
	}
	if structType.Fields == nil {
		return
	}
	for _, field := range structType.Fields.List {
		if len(field.Names) != 1 {
			continue
		}
		doc := field.Doc.Text()
		if doc != "" {
			descriptions[field.Names[0].Name] = strings.TrimSpace(doc)
		}
	}
}

func getTokenFromPulumiDirective(typ *doc.Type) (string, error) {
	const errorMsg = "resource type missing token; add a //pulumi:package:module:resource directive to the struct"
	if typ.Decl.Doc == nil || typ.Decl.Doc.List == nil {
		return "", errors.New(errorMsg)
	}
	for _, comment := range typ.Decl.Doc.List {
		for _, prefix := range []string{"//pulumi:", "// pulumi:"} {
			if strings.HasPrefix(comment.Text, prefix) {
				return strings.TrimPrefix(comment.Text, prefix), nil
			}
		}
	}
	return "", errors.New(errorMsg)
}

func rawMessage(v interface{}) json.RawMessage {
	bytes, err := json.Marshal(v)
	contract.Assert(err == nil)
	return bytes
}

func camel(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	res := make([]rune, 0, len(runes))
	for i, r := range runes {
		if unicode.IsLower(r) {
			res = append(res, runes[i:]...)
			break
		}
		res = append(res, unicode.ToLower(r))
	}
	return string(res)
}
