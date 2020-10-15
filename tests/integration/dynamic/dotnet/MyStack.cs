// Copyright 2016-2020, Pulumi Corporation.  All rights reserved.

using System;
using System.Security.Cryptography;
using System.Threading.Tasks;
using Pulumi;
using Pulumi.Dynamic;

class RandomResourceProvider : Pulumi.Dynamic.ResourceProvider
{
    public override Task<CreateResult> CreateAsync(object inputs)
    {
        var data = new byte[15];
        using var crypto = new RNGCryptoServiceProvider();
        crypto.GetBytes(data);
        string value = BitConverter.ToString(data).Replace("-", "");

        return Task.FromResult(new CreateResult{
            Id = value,
            Outputs =
            {
                { "val", value }
            },
        });
    }
}

class Random : Pulumi.Dynamic.Resource
{
    [Output("val")]
    public Output<string> Value { get; set; }

    public Random(string name, CustomResourceOptions? options = null)
        : base(new RandomResourceProvider(), name, RandomArgs.Instance, options)
    {
    }

    private sealed class RandomArgs : ResourceArgs
    {
        public static readonly RandomArgs Instance = new RandomArgs();

        [Input("val")]
        public readonly Input<string> Value = "";
    }
}

class MyStack : Stack
{
    [Output("random_id")]
    public Output<string> RandomId { get; set; }

    [Output("random_val")]
    public Output<string> RandomValue { get; set; }

    public MyStack()
    {
        var random = new Random("foo");
        RandomId = random.Id;
        RandomValue = random.Value;
    }
}
