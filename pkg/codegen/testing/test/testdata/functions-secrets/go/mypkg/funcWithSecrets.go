// Code generated by test DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package mypkg

import (
	"context"
	"reflect"

	"functions-secrets/mypkg/internal"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumix"
)

func FuncWithSecrets(ctx *pulumi.Context, args *FuncWithSecretsArgs, opts ...pulumi.InvokeOption) (*FuncWithSecretsResult, error) {
	opts = internal.PkgInvokeDefaultOpts(opts)
	var rv FuncWithSecretsResult
	err := ctx.Invoke("mypkg::funcWithSecrets", args, &rv, opts...)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

type FuncWithSecretsArgs struct {
	CryptoKey string `pulumi:"cryptoKey"`
	Plaintext string `pulumi:"plaintext"`
}

type FuncWithSecretsResult struct {
	Ciphertext string `pulumi:"ciphertext"`
	CryptoKey  string `pulumi:"cryptoKey"`
	Id         string `pulumi:"id"`
	Plaintext  string `pulumi:"plaintext"`
}

func FuncWithSecretsOutput(ctx *pulumi.Context, args FuncWithSecretsOutputArgs, opts ...pulumi.InvokeOption) FuncWithSecretsResultOutput {
	return pulumi.ToOutputWithContext(context.Background(), args).
		ApplyT(func(v interface{}) (FuncWithSecretsResult, error) {
			args := v.(FuncWithSecretsArgs)
			r, err := FuncWithSecrets(ctx, &args, opts...)
			var s FuncWithSecretsResult
			if r != nil {
				s = *r
			}
			return s, err
		}).(FuncWithSecretsResultOutput)
}

type FuncWithSecretsOutputArgs struct {
	CryptoKey pulumi.StringInput `pulumi:"cryptoKey"`
	Plaintext pulumi.StringInput `pulumi:"plaintext"`
}

func (FuncWithSecretsOutputArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*FuncWithSecretsArgs)(nil)).Elem()
}

type FuncWithSecretsResultOutput struct{ *pulumi.OutputState }

func (FuncWithSecretsResultOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*FuncWithSecretsResult)(nil)).Elem()
}

func (o FuncWithSecretsResultOutput) ToFuncWithSecretsResultOutput() FuncWithSecretsResultOutput {
	return o
}

func (o FuncWithSecretsResultOutput) ToFuncWithSecretsResultOutputWithContext(ctx context.Context) FuncWithSecretsResultOutput {
	return o
}

func (o FuncWithSecretsResultOutput) ToOutput(ctx context.Context) pulumix.Output[FuncWithSecretsResult] {
	return pulumix.Output[FuncWithSecretsResult]{
		OutputState: o.OutputState,
	}
}

func (o FuncWithSecretsResultOutput) Ciphertext() pulumi.StringOutput {
	return o.ApplyT(func(v FuncWithSecretsResult) string { return v.Ciphertext }).(pulumi.StringOutput)
}

func (o FuncWithSecretsResultOutput) CryptoKey() pulumi.StringOutput {
	return o.ApplyT(func(v FuncWithSecretsResult) string { return v.CryptoKey }).(pulumi.StringOutput)
}

func (o FuncWithSecretsResultOutput) Id() pulumi.StringOutput {
	return o.ApplyT(func(v FuncWithSecretsResult) string { return v.Id }).(pulumi.StringOutput)
}

func (o FuncWithSecretsResultOutput) Plaintext() pulumi.StringOutput {
	return o.ApplyT(func(v FuncWithSecretsResult) string { return v.Plaintext }).(pulumi.StringOutput)
}

func init() {
	pulumi.RegisterOutputType(FuncWithSecretsResultOutput{})
}
