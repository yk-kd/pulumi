// Copyright 2016-2020, Pulumi Corporation

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;

namespace Pulumi.Dynamic
{
    /// <summary>
    /// ResourceProvider is a Dynamic Resource Provider which allows defining new kinds of resources
    /// whose CRUD operations are implemented inside your .NET program.
    /// </summary>
    public abstract partial class ResourceProvider
    {
        protected ResourceProvider()
        {
        }

        //     def check(self, _olds: Any, news: Any) -> CheckResult:
        //         """
        //         Check validates that the given property bag is valid for a resource of the given type.
        //         """
        //         return CheckResult(news, [])
        /// <summary>
        /// Check validates that the given property bag is valid for a resource of the given type.
        /// </summary>
        public virtual Task<CheckResult> CheckAsync(object olds, object news)
            => Task.FromResult(new CheckResult());

        // /**
        // * Diff checks what impacts a hypothetical update will have on the resource's properties.
        // *
        // * @param id The ID of the resource to diff.
        // * @param olds The old values of properties to diff.
        // * @param news The new values of properties to diff.
        // */
        // diff?: (id: resource.ID, olds: any, news: any) => Promise<DiffResult>;


        //     def diff(self, _id: str, _olds: Any, _news: Any) -> DiffResult:
        //         """
        //         Diff checks what impacts a hypothetical update will have on the resource's properties.
        //         """
        //         return DiffResult()
        public virtual Task<DiffResult> DiffAsync(string id, object olds, object news)
            => Task.FromResult(new DiffResult());


        /// <summary>
        /// Create allocates a new instance of the provided resource and returns its unique ID afterwards.
        /// If this call fails, the resource must not have been created (i.e., it is "transactional").
        /// </summary>
        /// <param name="inputs">The properties to set during creation.</param>
        public abstract Task<CreateResult> CreateAsync(ImmutableDictionary<string, object> inputs);

        // /**
        // * Reads the current live state associated with a resource.  Enough state must be included in the inputs to uniquely
        // * identify the resource; this is typically just the resource ID, but it may also include some properties.
        // */
        // read?: (id: resource.ID, props?: any) => Promise<ReadResult>;

        //     def read(self, id_: str, props: Any) -> ReadResult:
        //         """
        //         Reads the current live state associated with a resource.  Enough state must be included in
        //         the inputs to uniquely identify the resource; this is typically just the resource ID, but it
        //         may also include some properties.
        //         """
        //         return ReadResult(id_, props)
        public virtual Task<ReadResult> ReadAsync(string id, object props)
            => Task.FromResult(new ReadResult());

        // /**
        // * Update updates an existing resource with new values.
        // *
        // * @param id The ID of the resource to update.
        // * @param olds The old values of properties to update.
        // * @param news The new values of properties to update.
        // */
        // update?: (id: resource.ID, olds: any, news: any) => Promise<UpdateResult>;

        //     def update(self, _id: str, _olds: Any, _news: Any) -> UpdateResult:
        //         """
        //         Update updates an existing resource with new values.
        //         """
        //         return UpdateResult()
        public virtual Task<UpdateResult> UpdateAsync(string id, object olds, object news)
            => Task.FromResult(new UpdateResult());

        // /**
        // * Delete tears down an existing resource with the given ID.  If it fails, the resource is assumed to still exist.
        // *
        // * @param id The ID of the resource to delete.
        // * @param props The current properties on the resource.
        // */
        // delete?: (id: resource.ID, props: any) => Promise<void>;

        //     def delete(self, _id: str, _props: Any) -> None:
        //         """
        //         Delete tears down an existing resource with the given ID.  If it fails, the resource is
        //         assumed to still exist.
        //         """
        public virtual Task DeleteAsync(string id, object props)
            => Task.CompletedTask;
    }

    public sealed class CheckResult
    {
    }

    public sealed class CheckFailure
    {
    }

    public sealed class DiffResult
    {
    }

    /// <summary>
    /// CreateResult represents the results of a call to <see cref="ResourceProvider.CreateAsync"/>.
    /// </summary>
    public sealed class CreateResult
    {
        /// <summary>
        /// The ID of the created resource.
        /// </summary>
        public string? Id { get; set; }

        private Dictionary<string, object>? _outputs;

        /// <summary>
        /// Any properties that were computed during creation.
        /// </summary>
        public Dictionary<string, object> Outputs
        {
            get => _outputs ??= new Dictionary<string, object>();
            set => _outputs = value;
        }
    }

    public sealed class ReadResult
    {
        // /**
        // * The ID of the resource ready back (or blank if missing).
        // */
        // readonly id?: resource.ID;
        // /**
        // * The current property state read from the live environment.
        // */
        // readonly props?: any;
    }

    /// <summary>
    /// UpdateResult represents the results of a call to `ResourceProvider.update`.
    /// </summary>
    public sealed class UpdateResult
    {
        // /**
        // * Any properties that were computed during updating.
        // */
        // readonly outs?: any;
    }
}
