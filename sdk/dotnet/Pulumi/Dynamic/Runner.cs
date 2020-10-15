// Copyright 2016-2020, Pulumi Corporation

using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Pulumi.Dynamic
{
    internal sealed class Runner : IRunner
    {
        public void RegisterTask(string description, Task task)
        {
            // Do nothing.
        }

        public Task<int> RunAsync(Func<Task<IDictionary<string, object?>>> func, StackOptions? options) => RunAsync();
        public Task<int> RunAsync<TStack>() where TStack : Stack, new() => RunAsync();

        private static Task<int> RunAsync()
        {
            Console.WriteLine("JVP Hello World!");
            return Task.FromResult(0);
        }
    }
}
