// Copyright 2016-2020, Pulumi Corporation

using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Reflection;
using System.Threading.Tasks;
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
                Ports = { new ServerPort("0.0.0.0", ServerPort.PickUnused, ServerCredentials.Insecure) }
            };
            int boundPort = server.Ports.Single().BoundPort;
            server.Start();

            Console.WriteLine(boundPort);
            Console.ReadLine();

            return Task.FromResult(0);
        }
    }

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
        private static ResourceProvider? GetProvider(Google.Protobuf.WellKnownTypes.Struct properties)
        {
            var serializedProvider = properties.Fields[Constants.ProviderPropertyName].StringValue;
            Debug.Assert(!string.IsNullOrEmpty(serializedProvider));
            string[] parts = serializedProvider.Split(':');
            Debug.Assert(parts.Length == 2);
            string typeFullName = parts[0];
            string brotliBase64 = parts[1];

            string path = Assembly.GetExecutingAssembly().Location;
            Assembly assembly = ResourceProvider.LoadFromBrotliBase64String(brotliBase64, path);

            Type? type = assembly.GetType(typeFullName, throwOnError: true);
            Debug.Assert(type != null);
            return Activator.CreateInstance(type) as ResourceProvider;
        }


        /// <summary>
        /// Diff checks what impacts a hypothetical update will have on the resource's properties.
        /// </summary>
        /// <param name="request">The request received from the client.</param>
        /// <param name="context">The context of the server-side call handler being invoked.</param>
        /// <returns>The response to send back to the client (wrapped by a task).</returns>
        public override Task<DiffResponse> Diff(DiffRequest request, ServerCallContext context)
        {
            Q.WriteLine("Diff");

            throw new RpcException(new Status(StatusCode.Unimplemented, ""));
        }

        /// <summary>
        /// Update updates an existing resource with new values.
        /// </summary>
        /// <param name="request">The request received from the client.</param>
        /// <param name="context">The context of the server-side call handler being invoked.</param>
        /// <returns>The response to send back to the client (wrapped by a task).</returns>
        public override Task<UpdateResponse> Update(UpdateRequest request, ServerCallContext context)
        {
            Q.WriteLine("Update");

            throw new RpcException(new Status(StatusCode.Unimplemented, ""));
        }


        /// <summary>
        /// Delete tears down an existing resource with the given ID.  If it fails, the resource is assumed to still exist.
        /// </summary>
        /// <param name="request">The request received from the client.</param>
        /// <param name="context">The context of the server-side call handler being invoked.</param>
        /// <returns>The response to send back to the client (wrapped by a task).</returns>
        public override Task<Google.Protobuf.WellKnownTypes.Empty> Delete(DeleteRequest request, ServerCallContext context)
        {
            Q.WriteLine("Delete");
            throw new RpcException(new Status(StatusCode.Unimplemented, ""));
        }

        /// <summary>
        /// Cancel signals the provider to abort all outstanding resource operations.
        /// </summary>
        /// <param name="request">The request received from the client.</param>
        /// <param name="context">The context of the server-side call handler being invoked.</param>
        /// <returns>The response to send back to the client (wrapped by a task).</returns>
        public override Task<Google.Protobuf.WellKnownTypes.Empty> Cancel(Google.Protobuf.WellKnownTypes.Empty request, ServerCallContext context)
            => Task.FromResult(new Google.Protobuf.WellKnownTypes.Empty());

        /// <summary>
        /// Create allocates a new instance of the provided resource and returns its unique ID afterwards.  (The input ID
        /// must be blank.)  If this call fails, the resource must not have been created (i.e., it is "transactional").
        /// </summary>
        /// <param name="request">The request received from the client.</param>
        /// <param name="context">The context of the server-side call handler being invoked.</param>
        /// <returns>The response to send back to the client (wrapped by a task).</returns>
        public override async Task<CreateResponse> Create(CreateRequest request, ServerCallContext context)
        {
            Q.WriteLine("Create");

            // props = rpc.deserialize_properties(request.properties)
            // provider = get_provider(props)
            // result = provider.create(props)
            // outs = result.outs
            // outs[PROVIDER_KEY] = props[PROVIDER_KEY]

            // loop = asyncio.new_event_loop()
            // outs_proto = loop.run_until_complete(rpc.serialize_properties(outs, {}))
            // loop.close()

            // fields = {"id": result.id, "properties": outs_proto}
            // return proto.CreateResponse(**fields)


            ResourceProvider? provider = GetProvider(request.Properties);
            Debug.Assert(provider != null);
            CreateResult result = await provider.CreateAsync(new object()).ConfigureAwait(false);

            var outs = new Google.Protobuf.WellKnownTypes.Struct();
            foreach (var (key, value) in result.Outputs) {
                // TODO fix all this
                outs.Fields.Add(key, new Google.Protobuf.WellKnownTypes.Value { StringValue = (string)value! });
            }
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

        /// <summary>
        /// Configure configures the resource provider with "globals" that control its behavior.
        /// </summary>
        /// <param name="request">The request received from the client.</param>
        /// <param name="context">The context of the server-side call handler being invoked.</param>
        /// <returns>The response to send back to the client (wrapped by a task).</returns>
        public override Task<ConfigureResponse> Configure(ConfigureRequest request, ServerCallContext context)
            => Task.FromResult(new ConfigureResponse { AcceptSecrets = false });

        /// <summary>
        /// GetPluginInfo returns generic information about this plugin, like its version.
        /// </summary>
        /// <param name="request">The request received from the client.</param>
        /// <param name="context">The context of the server-side call handler being invoked.</param>
        /// <returns>The response to send back to the client (wrapped by a task).</returns>
        public override Task<PluginInfo> GetPluginInfo(Google.Protobuf.WellKnownTypes.Empty request, ServerCallContext context)
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

            throw new RpcException(new Status(StatusCode.Unimplemented, ""));
        }
    }
}
