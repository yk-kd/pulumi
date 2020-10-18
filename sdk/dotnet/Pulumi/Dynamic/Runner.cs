// Copyright 2016-2020, Pulumi Corporation

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Diagnostics;
using System.Linq;
using System.Reflection;
using System.Threading.Tasks;
using Google.Protobuf.WellKnownTypes;
using Grpc.Core;
using Pulumi.Serialization;
using Pulumirpc;

namespace Pulumi.Dynamic
{
    internal sealed class Runner : IRunner
    {
        public void RegisterTask(string description, Task task) { }
        public Task<int> RunAsync(Func<Task<IDictionary<string, object?>>> func, StackOptions? options) => RunAsync();
        public Task<int> RunAsync<TStack>() where TStack : Stack, new() => RunAsync();

        private static Task<int> RunAsync()
        {
            // MaxRPCMessageSize raises the gRPC Max Message size from `4194304` (4mb) to `419430400` (400mb).
            const int MaxRPCMessageSize = 1024 * 1024 * 400;
            var server = new Server(new[] { new ChannelOption(ChannelOptions.MaxReceiveMessageLength, MaxRPCMessageSize) })
            {
                Services = { Pulumirpc.ResourceProvider.BindService(new ResourceProviderService()) },
                Ports = { new ServerPort("localhost", ServerPort.PickUnused, ServerCredentials.Insecure) }
            };
            int boundPort = server.Ports.Single().BoundPort;
            server.Start();

            Console.WriteLine(boundPort);
            Console.ReadLine();

            return Task.FromResult(0);
        }
    }

    // TODO remove
    internal static class Q
    {
        public static void WriteLine<T>(T value)
        {
            string tempDir = Environment.GetEnvironmentVariable("TMPDIR") ?? "/tmp";
            string path = System.IO.Path.Combine(tempDir, "q");
            using var writer = System.IO.File.AppendText(path);
            writer.WriteLine("{0}: {1}", DateTime.Now, value);
        }
    }

    internal sealed class ResourceProviderService : Pulumirpc.ResourceProvider.ResourceProviderBase
    {
        private static ResourceProvider GetProvider(Struct properties)
        {
            var serializedProvider = properties.Fields[Constants.ProviderPropertyName].StringValue;
            Debug.Assert(!string.IsNullOrEmpty(serializedProvider), "!string.IsNullOrEmpty(serializedProvider)");
            string[] parts = serializedProvider.Split(':');
            Debug.Assert(parts.Length == 2, "parts.Length == 2");
            string typeFullName = parts[0];
            string brotliBase64 = parts[1];

            string path = Assembly.GetExecutingAssembly().Location;
            Assembly assembly = ResourceProvider.LoadFromBrotliBase64String(brotliBase64, path);

            System.Type? type = assembly.GetType(typeFullName, throwOnError: true);
            Debug.Assert(type != null, "type != null");
            object? instance = Activator.CreateInstance(type);
            Debug.Assert(instance != null, "instance != null");
            return (ResourceProvider)instance;
        }

        public override async Task<DiffResponse> Diff(DiffRequest request, ServerCallContext context)
        {
            ImmutableDictionary<string, object> olds = ToDictionary(request.Olds);
            ImmutableDictionary<string, object> news = ToDictionary(request.News);

            ResourceProvider provider = request.News.Fields.ContainsKey(Constants.ProviderPropertyName)
                ? GetProvider(request.News)
                : GetProvider(request.Olds);

            DiffResult result = await provider.DiffAsync(request.Id, olds, news).ConfigureAwait(false);

            var response = new DiffResponse
            {
                Changes =
                    result.Changes == null ? DiffResponse.Types.DiffChanges.DiffUnknown :
                    result.Changes.Value ? DiffResponse.Types.DiffChanges.DiffSome :
                    DiffResponse.Types.DiffChanges.DiffNone,
            };
            response.Replaces.AddRange(result.Replaces);
            response.Stables.AddRange(result.Stables);
            if (result.DeleteBeforeReplace != null)
            {
                response.DeleteBeforeReplace = result.DeleteBeforeReplace.Value;
            }
            return response;
        }

        public override async Task<UpdateResponse> Update(UpdateRequest request, ServerCallContext context)
        {
            ImmutableDictionary<string, object> olds = ToDictionary(request.Olds);
            ImmutableDictionary<string, object> news = ToDictionary(request.News);

            ResourceProvider provider = GetProvider(request.News);

            UpdateResult result = await provider.UpdateAsync(request.Id, olds, news).ConfigureAwait(false);

            Struct outs = result.Outputs != null
                ? await SerializeAsync(result.Outputs).ConfigureAwait(false)
                : new Struct();
            outs.Fields.Add(Constants.ProviderPropertyName, request.News.Fields[Constants.ProviderPropertyName]);

            return new UpdateResponse { Properties = outs };
        }

        public override async Task<Empty> Delete(DeleteRequest request, ServerCallContext context)
        {
            ImmutableDictionary<string, object> props = ToDictionary(request.Properties);

            ResourceProvider provider = GetProvider(request.Properties);
            await provider.DeleteAsync(request.Id, props).ConfigureAwait(false);

            return new Empty();
        }

        public override Task<Empty> Cancel(Empty request, ServerCallContext context)
            => Task.FromResult(new Empty());

        public override async Task<CreateResponse> Create(CreateRequest request, ServerCallContext context)
        {
            ImmutableDictionary<string, object> props = ToDictionary(request.Properties);

            ResourceProvider provider = GetProvider(request.Properties);
            CreateResult result = await provider.CreateAsync(props).ConfigureAwait(false);

            Struct outs = result.Outputs != null
                ? await SerializeAsync(result.Outputs).ConfigureAwait(false)
                : new Struct();
            outs.Fields.Add(Constants.ProviderPropertyName, request.Properties.Fields[Constants.ProviderPropertyName]);

            return new CreateResponse
            {
                Id = result.Id,
                Properties = outs,
            };
        }

        /// <summary>
        /// Check validates that the given property bag is valid for a resource of the given type and returns the inputs
        /// that should be passed to successive calls to Diff, Create, or Update for this resource. As a rule, the provider
        /// inputs returned by a call to Check should preserve the original representation of the properties as present in
        /// the program inputs. Though this rule is not required for correctness, violations thereof can negatively impact
        /// the end-user experience, as the provider inputs are using for detecting and rendering diffs.
        /// </summary>
        /// <param name="request">The request received from the client.</param>
        /// <param name="context">The context of the server-side call handler being invoked.</param>
        /// <returns>The response to send back to the client (wrapped by a task).</returns>
        public override Task<CheckResponse> Check(CheckRequest request, ServerCallContext context)
        {
            Q.WriteLine("JVP CHECK CHECK CHECK");
            return Task.FromResult(new CheckResponse
            {
                Inputs = request.News
                // TODO Failures
            });
        }

        public override Task<ConfigureResponse> Configure(ConfigureRequest request, ServerCallContext context)
            => Task.FromResult(new ConfigureResponse { AcceptSecrets = false });

        public override Task<PluginInfo> GetPluginInfo(Empty request, ServerCallContext context)
            => Task.FromResult(new PluginInfo { Version = "0.1.0" });

        /// <summary>
        /// Read the current live state associated with a resource.  Enough state must be include in the inputs to uniquely
        /// identify the resource; this is typically just the resource ID, but may also include some properties.
        /// </summary>
        /// <param name="request">The request received from the client.</param>
        /// <param name="context">The context of the server-side call handler being invoked.</param>
        /// <returns>The response to send back to the client (wrapped by a task).</returns>
        public override Task<ReadResponse> Read(ReadRequest request, ServerCallContext context)
        {
            Q.WriteLine("Read");

            // id_ = request.id
            // props = rpc.deserialize_properties(request.properties)
            // provider = get_provider(props)
            // result = provider.read(id_, props)
            // outs = result.outs
            // outs[PROVIDER_KEY] = props[PROVIDER_KEY]

            // loop = asyncio.new_event_loop()
            // outs_proto = loop.run_until_complete(rpc.serialize_properties(outs, {}))
            // loop.close()

            // fields = {"id": result.id, "properties": outs_proto}
            // return proto.ReadResponse(**fields)

            string id = request.Id;
            var outs = request.Properties;
            //var outs = new Struct();
            //outs.Fields.Add(Constants.ProviderPropertyName, request.Properties.Fields[Constants.ProviderPropertyName]);
            return Task.FromResult(new ReadResponse
            {
                Id = id,
                Properties = outs,
            });
        }

        // TODO dedupe ToDictionary and SerializeAsync from MockMonitor.cs.
        private static ImmutableDictionary<string, object> ToDictionary(Struct s)
        {
            var builder = ImmutableDictionary.CreateBuilder<string, object>();
            foreach (var (key, value) in s.Fields)
            {
                var data = Deserializer.Deserialize(value);
                if (data.IsKnown && data.Value != null)
                {
                    builder.Add(key, data.Value);
                }
            }
            return builder.ToImmutable();
        }

        private async Task<Struct> SerializeAsync(object o)
        {
            var dict = (o as IDictionary<string, object>)?.ToImmutableDictionary()
                       ?? await new Serializer().SerializeAsync("", o).ConfigureAwait(false) as ImmutableDictionary<string, object>
                       ?? throw new InvalidOperationException($"{o.GetType().FullName} is not a supported argument type");
            return Serializer.CreateStruct(dict);
        }
    }
}
