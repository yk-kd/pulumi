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

        /// <summary>
        /// Diff checks what impacts a hypothetical update will have on the resource's properties.
        /// </summary>
        /// <param name="id">The ID of the resource to diff.</param>
        /// <param name="olds">The old values of properties to diff.</param>
        /// <param name="news">The new values of properties to diff.</param>
        public virtual Task<DiffResult> DiffAsync(string id, ImmutableDictionary<string, object> olds, ImmutableDictionary<string, object> news)
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

        /// <summary>
        /// Update updates an existing resource with new values.
        /// </summary>
        /// <param name="id">The ID of the resource to update.</param>
        /// <param name="olds">The old values of properties to update.</param>
        /// <param name="news">The new values of properties to update.</param>
        public virtual Task<UpdateResult> UpdateAsync(string id, ImmutableDictionary<string, object> olds, ImmutableDictionary<string, object> news)
            => Task.FromResult(new UpdateResult());

        /// <summary>
        /// Delete tears down an existing resource with the given ID.
        /// If it fails, the resource is assumed to still exist.
        /// </summary>
        /// <param name="id">The ID of the resource to delete.</param>
        /// <param name="props">The current properties on the resource.</param>
        public virtual Task DeleteAsync(string id, ImmutableDictionary<string, object> props)
            => Task.CompletedTask;
    }

    public sealed class CheckResult
    {
    }

    public sealed class CheckFailure
    {
    }

    /// <summary>
    /// DiffResult represents the results of a call to <see cref="ResourceProvider.DiffAsync"/>.
    /// </summary>
    public sealed class DiffResult
    {
        /// <summary>
        /// If true, this diff detected changes and suggests an update.
        /// </summary>
        public bool? Changes { get; set; }

        private List<string>? _replaces;

        /// <summary>
        /// If this update requires a replacement, the set of properties triggering it.
        /// </summary>
        public List<string> Replaces
        {
            get => _replaces ??= new List<string>();
            set => _replaces = value;
        }

        private List<string>? _stables;

        /// <summary>
        /// An optional list of properties that will not ever change.
        /// </summary>
        public List<string> Stables
        {
            get => _stables ??= new List<string>();
            set => _stables = value;
        }

        /// <summary>
        /// If true, and a replacement occurs, the resource will first be deleted before being recreated.
        /// This is to void potential side-by-side issues with the default create before delete behavior.
        /// </summary>
        public bool? DeleteBeforeReplace { get; set; }
    }

    /// <summary>
    /// CreateResult represents the results of a call to <see cref="ResourceProvider.CreateAsync"/>.
    /// </summary>
    public sealed class CreateResult
    {
        /// <summary>
        /// The ID of the created resource.
        /// </summary>
        public string Id { get; set; } = null!;

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
    /// UpdateResult represents the results of a call to <see cref="ResourceProvider.UpdateAsync"/>.
    /// </summary>
    public sealed class UpdateResult
    {
        private Dictionary<string, object>? _outputs;

        /// <summary>
        /// Any properties that were computed during updating.
        /// </summary>
        public Dictionary<string, object> Outputs
        {
            get => _outputs ??= new Dictionary<string, object>();
            set => _outputs = value;
        }
    }
}
