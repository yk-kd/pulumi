// *** WARNING: this file was generated by test. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

import * as pulumiRandom from "@pulumi/random";

export function argFunction(args?: ArgFunctionArgs, opts?: pulumi.InvokeOptions): Promise<ArgFunctionResult> {
    args = args || {};
    opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts || {});
    return pulumi.runtime.invoke("example::argFunction", {
        "name": args.name,
    }, opts);
}

export interface ArgFunctionArgs {
    name?: pulumiRandom.RandomPet;
}

export interface ArgFunctionResult {
    readonly age?: number;
}


export function argFunctionOutput(args?: ArgFunctionOutputArgs, opts?: pulumi.InvokeOptions): pulumi.Output<ArgFunctionResult> {
    return pulumi.output(args).apply((a: any) => argFunction(a, opts))
}
export interface ArgFunctionOutputArgs {
    name?: pulumi.Input<pulumiRandom.RandomPet>;
}
