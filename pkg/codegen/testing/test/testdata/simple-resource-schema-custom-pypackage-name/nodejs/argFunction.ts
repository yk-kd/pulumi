// *** WARNING: this file was generated by test. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

import {Resource} from "./index";

export function argFunction(args?: ArgFunctionArgs, opts?: pulumi.InvokeOptions): Promise<ArgFunctionResult> {
    args = args || {};
    opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts || {});
    return pulumi.runtime.invoke("example::argFunction", {
        "arg1": args.arg1,
    }, opts);
}

export interface ArgFunctionArgs {
    arg1?: Resource;
}

export interface ArgFunctionResult {
    readonly result?: Resource;
}


export function argFunctionOutput(args?: ArgFunctionOutputArgs, opts?: pulumi.InvokeOptions): pulumi.Output<ArgFunctionResult> {
    return pulumi.output(args).apply((a: any) => argFunction(a, opts))
}
export interface ArgFunctionOutputArgs {
    arg1?: pulumi.Input<Resource>;
}
