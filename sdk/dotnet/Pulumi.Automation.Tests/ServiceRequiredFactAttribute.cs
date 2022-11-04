// Copyright 2016-2022, Pulumi Corporation

using System;
using Xunit;

namespace Pulumi.Automation.Tests
{
    public sealed class ServiceRequiredFactAttribute : FactAttribute
    {
        public ServiceRequiredFactAttribute()
        {
            if (Environment.GetEnvironmentVariable("PULUMI_ACCESS_TOKEN") is null)
            {
                Skip = "PULUMI_ACCESS_TOKEN not set";
            }
        }
    }
}
