// Copyright 2016-2021, Pulumi Corporation

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Linq;
using System.Threading.Tasks;
using Pulumi.Testing;
using Xunit;

namespace Pulumi.Tests.Mocks
{
    public class Issue7422
    {
        class Mocks : IMocks
        {
            public Task<object> CallAsync(MockCallArgs args)
            {
                return Task.FromResult<object>(args);
            }

            public Task<(string? id, object state)> NewResourceAsync(MockResourceArgs args) =>
                args.Type switch
                {
                    "pkg:index:MyCustom" => Task.FromResult<(string?, object)>((args.Name + "_id", args.Inputs.Add("urn", "some_non_pulumi_urn"))),
                    _ => throw new Exception($"Unknown resource {args.Type}")
                };
        }

        public sealed class MyCustomArgs : ResourceArgs
        {
        }

        public class MyCustom : CustomResource
        {
            [Output("urn")]
            public new Output<string> Urn { get; private set; } = null!;

            public MyCustom(string name, MyCustomArgs args, CustomResourceOptions? options = null)
                : base("pkg:index:MyCustom", name, args, options)
            {
            }
        }

        [Fact]
        public async Task Test()
        {
            var (resources, exception) = await Testing.Run(new Mocks(), () => {
                var myCustom = new MyCustom("name", new MyCustomArgs());
                return new Dictionary<string, object?>(){};
            });

            var stack = resources.OfType<Stack>().FirstOrDefault();
            Assert.NotNull(stack);

            var resource = resources.OfType<MyCustom>().FirstOrDefault();
            Assert.NotNull(resource);
            Assert.Equal("urn:pulumi:stack::project::pulumi:pulumi:Stack$pkg:index:MyCustom::name", await resource.Urn.GetValueAsync("<UNKNOWN>"));
            Assert.Null((resource as CustomResource).Urn);

            Assert.NotNull(exception);
            Assert.StartsWith("Running program '", exception!.Message);
            Assert.Contains("' failed with an unhandled exception:", exception!.Message);
            Assert.Contains("System.Exception: Uninitialized `Urn` property on pkg:index:MyCustom:name resource.", exception!.Message);
        }
    }
}