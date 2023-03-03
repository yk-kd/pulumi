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
package codegen

import (
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	codegenrpc "github.com/pulumi/pulumi/sdk/v3/proto/go/codegen"
)

// Returns converter plugin specific data for the requested provider. This will normally be human readable JSON, but the engine doesn't mandate any form.
type GetMappingResponse struct {
	// The conversion plugin specific data (if any).
	Data []byte
}

func (s *GetMappingResponse) marshal() *codegenrpc.GetMappingResponse {
	return &codegenrpc.GetMappingResponse{
		Data: s.Data,
	}
}

func (s *GetMappingResponse) unmarshal(data *codegenrpc.GetMappingResponse) {
	s.Data = data.Data
}
