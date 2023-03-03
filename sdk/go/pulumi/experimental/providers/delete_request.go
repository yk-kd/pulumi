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
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	codegenrpc "github.com/pulumi/pulumi/sdk/v3/proto/go/codegen"
)

// A request to delete a resource.
type DeleteRequest struct {
	// The ID of the resource to delete.
	Id string
	// The Pulumi URN for this resource.
	URN string
	// The Pulumi type for this resource.
	Type string
	// The Pulumi name for this resource.
	Name string
}

func (s *DeleteRequest) marshal() *pulumirpc.DeleteRequest {
	return &pulumirpc.DeleteRequest{
		Id:  s.Id,
		Urn: s.URN,
	}
}

func (s *DeleteRequest) unmarshal(data *pulumirpc.DeleteRequest) {
	s.Id = data.Id
	s.URN = data.Urn
	s.Type = string(resource.URN(data.Urn).Type())
	s.Name = string(resource.URN(data.Urn).Name())
}
