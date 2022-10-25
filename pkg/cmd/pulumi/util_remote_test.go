// Copyright 2016-2022, Pulumi Corporation.
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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEnv(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     string
		name      string
		value     string
		secret    bool
		errString string
	}{
		"name val":                          {input: "FOO=bar", name: "FOO", value: "bar"},
		"name val secret":                   {input: "FOO#=bar", name: "FOO", value: "bar", secret: true},
		"name empty val":                    {input: "FOO=", name: "FOO"},
		"name empty val secret":             {input: "FOO#=", name: "FOO", secret: true},
		"name val extra seps":               {input: "FOO=bar=baz", name: "FOO", value: "bar=baz"},
		"name val extra seps secret":        {input: "FOO#=bar=baz", name: "FOO", value: "bar=baz", secret: true},
		"name val extra secret seps":        {input: "FOO=bar#=baz", name: "FOO", value: "bar#=baz"},
		"name val extra secret seps secret": {input: "FOO#=bar#=baz", name: "FOO", value: "bar#=baz", secret: true},
		"empty":                             {input: "", errString: `error parsing ""`},
		"no sep":                            {input: "foo", errString: `error parsing "foo"`},
		"just secret sep":                   {input: "#", errString: `error parsing "#"`},
		"empty name val":                    {input: "=", errString: `expected non-empty environment name for "="`},
		"empty secret name val":             {input: "#=", errString: `expected non-empty environment name for "#="`},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			name, value, secret, err := parseEnv(tc.input)
			if tc.errString != "" {
				assert.EqualError(t, err, tc.errString)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.name, name)
			assert.Equal(t, tc.value, value)
			assert.Equal(t, tc.secret, secret)
		})
	}
}
