// Copyright 2016-2020, Pulumi Corporation

using System;
using System.Threading.Tasks;

namespace Pulumi.Dynamic
{
    /// <summary>
    /// Resource represents a Pulumi Resource that incorporates an inline implementation of the Resource's CRUD operations.
    /// </summary>
    public abstract class Resource : CustomResource
    {
        protected Resource(ResourceProvider provider, string name, ResourceArgs args, CustomResourceOptions? options = null)
            : base("pulumi-dotnet:dynamic:Resource", name, ArgsWithProvider(provider, args), options)
        {
        }

        private static ResourceArgs ArgsWithProvider(ResourceProvider provider, ResourceArgs args)
        {
            if (provider is null)
            {
                throw new ArgumentNullException(nameof(provider));
            }

            string serialized = provider.Serialize();
            return (ResourceArgs)args.WithProvider($"{provider.GetType().FullName}:{serialized}");
        }
    }



    // export abstract class Resource extends resource.CustomResource {
    //     /**
    //     * Creates a new dynamic resource.
    //     *
    //     * @param provider The implementation of the resource's CRUD operations.
    //     * @param name The name of the resource.
    //     * @param props The arguments to use to populate the new resource. Must not define the reserved
    //     *              property "__provider".
    //     * @param opts A bag of options that control this resource's behavior.
    //     */
    //     constructor(provider: ResourceProvider, name: string, props: Inputs,
    //                 opts?: resource.CustomResourceOptions) {
    //         const providerKey: string = "__provider";
    //         if (props[providerKey]) {
    //             throw new Error("A dynamic resource must not define the __provider key");
    //         }
    //         props[providerKey] = serializeProvider(provider);

    //         super("pulumi-nodejs:dynamic:Resource", name, props, opts);
    //     }
    // }

}
