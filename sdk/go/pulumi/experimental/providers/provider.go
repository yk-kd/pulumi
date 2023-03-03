// Copyright 2016-2023, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by "generate"; DO NOT EDIT.

//nolint:lll
package providers

import (
	"context"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// Provider presents a simple interface for orchestrating resource create, read, update, and delete operations. Each provider understands how to handle all of the resource types within a single package.
//
// It is important to note that provider operations are not transactional (Some providers might decide to offer transactional semantics, but such a provider is a rare treat). As a result, failures in the operations below can range from benign to catastrophic (possibly leaving behind a corrupt resource). It is up to the provider to make a best effort to ensure catastrophes do not occur. The errors returned from mutating operations indicate both the underlying error condition in addition to a bit indicating whether the operation was successfully rolled back.
type Provider interface {
	// Validates the configuration for this resource provider.
	CheckConfig(ctx context.Context, request CheckRequest) (CheckResponse, error)
	// Checks what impacts a hypothetical change to this provider's configuration will have on the provider.
	DiffConfig(ctx context.Context, request DiffRequest) (DiffResponse, error)
	// Configures the resource provider with "globals" that control its behavior.
	Configure(ctx context.Context, request ConfigureRequest) (ConfigureResponse, error)
	// Validates that the given property bag is valid for a resource of the given type and returns the inputs that should be passed to successive calls to Diff, Create, or Update for this resource. As a rule, the provider inputs returned by a call to Check should preserve the original representation of the properties as present in the program inputs. Though this rule is not required for correctness, violations thereof can negatively impact the end-user experience, as the provider inputs are using for detecting and rendering diffs.
	Check(ctx context.Context, request CheckRequest) (CheckResponse, error)
	// Checks what impacts a hypothetical update will have on the resource's properties.
	Diff(ctx context.Context, request DiffRequest) (DiffResponse, error)
	// Allocates a new instance of the provided resource and returns its unique ID afterwards. (The input ID must be blank.)  If this call fails, the resource must not have been created (i.e., it is "transactional").
	Create(ctx context.Context, request CreateRequest) (CreateResponse, error)
	// Updates an existing resource with new values.
	Update(ctx context.Context, request UpdateRequest) (UpdateResponse, error)
	// Tears down an existing resource with the given ID. If it fails, the resource is assumed to still exist.
	Delete(ctx context.Context, request DeleteRequest) error
	// Reads the current live state associated with a resource. Enough state must be include in the inputs to uniquely identify the resource; this is typically just the resource ID, but may also include some properties.
	Read(ctx context.Context, request ReadRequest) (ReadResponse, error)
}
