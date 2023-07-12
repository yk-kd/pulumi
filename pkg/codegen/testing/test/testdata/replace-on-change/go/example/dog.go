// Code generated by test DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package example

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumix"
	"replace-on-change/example/internal"
)

type Dog struct {
	pulumi.CustomResourceState

	Bone pulumi.StringPtrOutput `pulumi:"bone"`
}

// NewDog registers a new resource with the given unique name, arguments, and options.
func NewDog(ctx *pulumi.Context,
	name string, args *DogArgs, opts ...pulumi.ResourceOption) (*Dog, error) {
	if args == nil {
		args = &DogArgs{}
	}

	replaceOnChanges := pulumi.ReplaceOnChanges([]string{
		"bone",
	})
	opts = append(opts, replaceOnChanges)
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource Dog
	err := ctx.RegisterResource("example::Dog", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetDog gets an existing Dog resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetDog(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *DogState, opts ...pulumi.ResourceOption) (*Dog, error) {
	var resource Dog
	err := ctx.ReadResource("example::Dog", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering Dog resources.
type dogState struct {
}

type DogState struct {
}

func (DogState) ElementType() reflect.Type {
	return reflect.TypeOf((*dogState)(nil)).Elem()
}

type dogArgs struct {
}

// The set of arguments for constructing a Dog resource.
type DogArgs struct {
}

func (DogArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*dogArgs)(nil)).Elem()
}

type DogInput interface {
	pulumi.Input

	ToDogOutput() DogOutput
	ToDogOutputWithContext(ctx context.Context) DogOutput
}

func (*Dog) ElementType() reflect.Type {
	return reflect.TypeOf((**Dog)(nil)).Elem()
}

func (i *Dog) ToDogOutput() DogOutput {
	return i.ToDogOutputWithContext(context.Background())
}

func (i *Dog) ToDogOutputWithContext(ctx context.Context) DogOutput {
	return pulumi.ToOutputWithContext(ctx, i).(DogOutput)
}

func (i *Dog) ToOutput(ctx context.Context) pulumix.Output[*Dog] {
	return pulumix.Output[*Dog]{
		OutputState: i.ToDogOutputWithContext(ctx).OutputState,
	}
}

// DogArrayInput is an input type that accepts DogArray and DogArrayOutput values.
// You can construct a concrete instance of `DogArrayInput` via:
//
//	DogArray{ DogArgs{...} }
type DogArrayInput interface {
	pulumi.Input

	ToDogArrayOutput() DogArrayOutput
	ToDogArrayOutputWithContext(context.Context) DogArrayOutput
}

type DogArray []DogInput

func (DogArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*Dog)(nil)).Elem()
}

func (i DogArray) ToDogArrayOutput() DogArrayOutput {
	return i.ToDogArrayOutputWithContext(context.Background())
}

func (i DogArray) ToDogArrayOutputWithContext(ctx context.Context) DogArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(DogArrayOutput)
}

func (i DogArray) ToOutput(ctx context.Context) pulumix.Output[[]*Dog] {
	return pulumix.Output[[]*Dog]{
		OutputState: i.ToDogArrayOutputWithContext(ctx).OutputState,
	}
}

// DogMapInput is an input type that accepts DogMap and DogMapOutput values.
// You can construct a concrete instance of `DogMapInput` via:
//
//	DogMap{ "key": DogArgs{...} }
type DogMapInput interface {
	pulumi.Input

	ToDogMapOutput() DogMapOutput
	ToDogMapOutputWithContext(context.Context) DogMapOutput
}

type DogMap map[string]DogInput

func (DogMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*Dog)(nil)).Elem()
}

func (i DogMap) ToDogMapOutput() DogMapOutput {
	return i.ToDogMapOutputWithContext(context.Background())
}

func (i DogMap) ToDogMapOutputWithContext(ctx context.Context) DogMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(DogMapOutput)
}

func (i DogMap) ToOutput(ctx context.Context) pulumix.Output[map[string]*Dog] {
	return pulumix.Output[map[string]*Dog]{
		OutputState: i.ToDogMapOutputWithContext(ctx).OutputState,
	}
}

type DogOutput struct{ *pulumi.OutputState }

func (DogOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**Dog)(nil)).Elem()
}

func (o DogOutput) ToDogOutput() DogOutput {
	return o
}

func (o DogOutput) ToDogOutputWithContext(ctx context.Context) DogOutput {
	return o
}

func (o DogOutput) ToOutput(ctx context.Context) pulumix.Output[*Dog] {
	return pulumix.Output[*Dog]{
		OutputState: o.OutputState,
	}
}

func (o DogOutput) Bone() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *Dog) pulumi.StringPtrOutput { return v.Bone }).(pulumi.StringPtrOutput)
}

type DogArrayOutput struct{ *pulumi.OutputState }

func (DogArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*Dog)(nil)).Elem()
}

func (o DogArrayOutput) ToDogArrayOutput() DogArrayOutput {
	return o
}

func (o DogArrayOutput) ToDogArrayOutputWithContext(ctx context.Context) DogArrayOutput {
	return o
}

func (o DogArrayOutput) ToOutput(ctx context.Context) pulumix.Output[[]*Dog] {
	return pulumix.Output[[]*Dog]{
		OutputState: o.OutputState,
	}
}

func (o DogArrayOutput) Index(i pulumi.IntInput) DogOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *Dog {
		return vs[0].([]*Dog)[vs[1].(int)]
	}).(DogOutput)
}

type DogMapOutput struct{ *pulumi.OutputState }

func (DogMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*Dog)(nil)).Elem()
}

func (o DogMapOutput) ToDogMapOutput() DogMapOutput {
	return o
}

func (o DogMapOutput) ToDogMapOutputWithContext(ctx context.Context) DogMapOutput {
	return o
}

func (o DogMapOutput) ToOutput(ctx context.Context) pulumix.Output[map[string]*Dog] {
	return pulumix.Output[map[string]*Dog]{
		OutputState: o.OutputState,
	}
}

func (o DogMapOutput) MapIndex(k pulumi.StringInput) DogOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *Dog {
		return vs[0].(map[string]*Dog)[vs[1].(string)]
	}).(DogOutput)
}

func init() {
	pulumi.RegisterOutputType(DogOutput{})
	pulumi.RegisterOutputType(DogArrayOutput{})
	pulumi.RegisterOutputType(DogMapOutput{})
}
