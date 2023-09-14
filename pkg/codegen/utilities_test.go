// Copyright 2016-2021, Pulumi Corporation.
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

package codegen

import (
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/stretchr/testify/assert"
)

func TestSimplifyInputUnion(t *testing.T) {
	t.Parallel()

	u1 := &schema.UnionType{
		ElementTypes: []schema.Type{
			&schema.InputType{ElementType: schema.StringType},
			schema.NumberType,
		},
	}

	u2 := SimplifyInputUnion(u1)
	assert.Equal(t, &schema.UnionType{
		ElementTypes: []schema.Type{
			schema.StringType,
			schema.NumberType,
		},
	}, u2)
}
