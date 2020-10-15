// Copyright 2016-2020, Pulumi Corporation

using System;
using System.IO;
using System.IO.Compression;
using System.Reflection;
using System.Runtime.Loader;

namespace Pulumi.Dynamic
{
    partial class ResourceProvider
    {
        internal string Serialize()
        {
            //string path = Assembly.GetExecutingAssembly().Location;
            string path = GetType().Assembly.Location;
            // TODO use the ILLinker to prune the assembly to just the provider and the code it depends on.
            return FileToBrotliBase64String(path);
        }

        internal static string FileToBrotliBase64String(string path)
        {
            using var memory = new MemoryStream();
            using (var file = File.OpenRead(path))
            using (var brotli = new BrotliStream(memory, CompressionMode.Compress))
            {
                file.CopyTo(brotli);
            }
            byte[] bytes = memory.ToArray();
            return Convert.ToBase64String(bytes);
        }

        internal static Assembly LoadFromBrotliBase64String(string value, string dependencyDirectory)
        {
            byte[] bytes = Convert.FromBase64String(value);
            using var memory = new MemoryStream();
            using (var source = new MemoryStream(bytes))
            using (var brotli = new BrotliStream(source, CompressionMode.Decompress))
            {
                brotli.CopyTo(memory);
            }
            memory.Position = 0;

            var loadContext = new ProviderLoadContext(dependencyDirectory);
            return loadContext.LoadFromStream(memory);

            //return AssemblyLoadContext.Default.LoadFromStream(memory);
        }

        internal sealed class ProviderLoadContext : AssemblyLoadContext
        {
            private readonly AssemblyDependencyResolver _resolver;

            public ProviderLoadContext(string path)
            {
                _resolver = new AssemblyDependencyResolver(path);
            }

            protected override Assembly? Load(AssemblyName assemblyName)
            {
                string? assemblyPath = _resolver.ResolveAssemblyToPath(assemblyName);
                return assemblyPath != null ? LoadFromAssemblyPath(assemblyPath) : null;
            }

            protected override IntPtr LoadUnmanagedDll(string unmanagedDllName)
            {
                string? libraryPath = _resolver.ResolveUnmanagedDllToPath(unmanagedDllName);
                return libraryPath != null ? LoadUnmanagedDllFromPath(libraryPath) : IntPtr.Zero;
            }
        }
    }
}
