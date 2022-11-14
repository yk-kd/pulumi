// *** WARNING: this file was generated by test. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

/**
 * Returns the absolute value of a given float.
 * Example: abs(1) returns 1, and abs(-1) would also return 1, whereas abs(-3.14) would return 3.14.
 */
export function absMultiArgsReducedOutputSwapped(b: number, a: number, opts?: pulumi.InvokeOptions): Promise<number> {
    opts = pulumi.mergeOptions(utilities.resourceOptsDefaults(), opts || {});
    return pulumi.runtime.invoke("std:index:AbsMultiArgsReducedOutputSwapped", {
        "b": b,
        "a": a,
    }, opts).then((r: any) => r.result);
}

/**
 * Returns the absolute value of a given float.
 * Example: abs(1) returns 1, and abs(-1) would also return 1, whereas abs(-3.14) would return 3.14.
 */
export function absMultiArgsReducedOutputSwappedOutput(b: pulumi.Input<number>, a: pulumi.Input<number>, opts?: pulumi.InvokeOptions): pulumi.Output<number> {
    var args = {
        "b": b,
        "a": a,
    };
    return pulumi.output(args).apply((resolvedArgs: any) => absMultiArgsReducedOutputSwapped(resolvedArgs.b, resolvedArgs.a, opts))
}
