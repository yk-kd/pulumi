// Copyright 2016-2020, Pulumi Corporation.
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
package plugin

import (
	"fmt"
	"testing"

	"github.com/pulumi/pulumi/pkg/apitype"
	"github.com/stretchr/testify/assert"
)

func TestParsePolicyPackConfigSuccess(t *testing.T) {
	tests := []struct {
		JSON     string
		Expected map[string]AnalyzerPolicyConfig
	}{
		{
			JSON:     `{}`,
			Expected: map[string]AnalyzerPolicyConfig{},
		},
		{
			JSON: `{"foo":"advisory"}`,
			Expected: map[string]AnalyzerPolicyConfig{
				"foo": AnalyzerPolicyConfig{
					EnforcementLevel: apitype.Advisory,
				},
			},
		},
		{
			JSON: `{"foo":"mandatory"}`,
			Expected: map[string]AnalyzerPolicyConfig{
				"foo": AnalyzerPolicyConfig{
					EnforcementLevel: apitype.Mandatory,
				},
			},
		},
		{
			JSON: `{"foo":{"enforcementLevel":"advisory"}}`,
			Expected: map[string]AnalyzerPolicyConfig{
				"foo": AnalyzerPolicyConfig{
					EnforcementLevel: apitype.Advisory,
				},
			},
		},
		{
			JSON: `{"foo":{"enforcementLevel":"mandatory"}}`,
			Expected: map[string]AnalyzerPolicyConfig{
				"foo": AnalyzerPolicyConfig{
					EnforcementLevel: apitype.Mandatory,
				},
			},
		},
		{
			JSON: `{"foo":{"enforcementLevel":"advisory","bar":"blah"}}`,
			Expected: map[string]AnalyzerPolicyConfig{
				"foo": AnalyzerPolicyConfig{
					EnforcementLevel: apitype.Advisory,
					Properties: map[string]interface{}{
						"bar": "blah",
					},
				},
			},
		},
		{
			JSON:     `{"foo":{}}`,
			Expected: map[string]AnalyzerPolicyConfig{},
		},
		{
			JSON: `{"foo":{"bar":"blah"}}`,
			Expected: map[string]AnalyzerPolicyConfig{
				"foo": AnalyzerPolicyConfig{
					Properties: map[string]interface{}{
						"bar": "blah",
					},
				},
			},
		},
		{
			JSON: `{"policy1":{"foo":"one"},"policy2":{"foo":"two"}}`,
			Expected: map[string]AnalyzerPolicyConfig{
				"policy1": AnalyzerPolicyConfig{
					Properties: map[string]interface{}{
						"foo": "one",
					},
				},
				"policy2": AnalyzerPolicyConfig{
					Properties: map[string]interface{}{
						"foo": "two",
					},
				},
			},
		},
		{
			JSON: `{"all":"mandatory","policy1":{"foo":"one"},"policy2":{"foo":"two"}}`,
			Expected: map[string]AnalyzerPolicyConfig{
				"all": AnalyzerPolicyConfig{
					EnforcementLevel: apitype.Mandatory,
				},
				"policy1": AnalyzerPolicyConfig{
					Properties: map[string]interface{}{
						"foo": "one",
					},
				},
				"policy2": AnalyzerPolicyConfig{
					Properties: map[string]interface{}{
						"foo": "two",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
			result, err := parsePolicyPackConfig([]byte(test.JSON))
			assert.NoError(t, err)
			assert.Equal(t, test.Expected, result)
		})
	}
}

func TestParsePolicyPackConfigFail(t *testing.T) {
	tests := []string{
		``,
		`{"foo":[]}`,
		`{"foo":null}`,
		`{"foo":undefined}`,
		`{"foo":0}`,
		`{"foo":true}`,
		`{"foo":false}`,
		`{"foo":""}`,
		`{"foo":"bar"}`,
		`{"foo":{"enforcementLevel":[]}}`,
		`{"foo":{"enforcementLevel":null}}`,
		`{"foo":{"enforcementLevel":undefined}}`,
		`{"foo":{"enforcementLevel":0}}`,
		`{"foo":{"enforcementLevel":true}}`,
		`{"foo":{"enforcementLevel":false}}`,
		`{"foo":{"enforcementLevel":{}}}`,
		`{"foo":{"enforcementLevel":""}}`,
		`{"foo":{"enforcementLevel":"bar"}}`,
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
			result, err := parsePolicyPackConfig([]byte(test))
			assert.Nil(t, result)
			assert.Error(t, err)
		})
	}
}

func TestCreateConfigWithDefaultsEnforcementLevel(t *testing.T) {
	tests := []struct {
		Policies []AnalyzerPolicyInfo
		Expected map[string]*AnalyzerPolicyConfig
	}{
		{
			Policies: []AnalyzerPolicyInfo{
				{
					Name:             "policy",
					EnforcementLevel: "advisory",
				},
			},
			Expected: map[string]*AnalyzerPolicyConfig{
				"policy": &AnalyzerPolicyConfig{
					EnforcementLevel: "advisory",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
			result, err := CreateConfigWithDefaults(test.Policies)
			assert.NoError(t, err)
			assert.Equal(t, test.Expected, result)
		})
	}
}

func TestCreateConfigWithDefaults(t *testing.T) {
	tests := []struct {
		Policies []AnalyzerPolicyInfo
		Expected map[string]*AnalyzerPolicyConfig
	}{
		{
			Policies: []AnalyzerPolicyInfo{
				{
					Name:             "policy",
					EnforcementLevel: "advisory",
					ConfigSchema: &AnalyzerPolicyConfigSchema{
						Properties: map[string]JSONSchema{
							"foo": JSONSchema{
								"type":    "string",
								"default": "bar",
							},
						},
					},
				},
			},
			Expected: map[string]*AnalyzerPolicyConfig{
				"policy": &AnalyzerPolicyConfig{
					EnforcementLevel: "advisory",
					Properties: map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
		{
			Policies: []AnalyzerPolicyInfo{
				{
					Name:             "policy",
					EnforcementLevel: "advisory",
					ConfigSchema: &AnalyzerPolicyConfigSchema{
						Properties: map[string]JSONSchema{
							"foo": JSONSchema{
								"type":    "number",
								"default": float64(42),
							},
						},
					},
				},
			},
			Expected: map[string]*AnalyzerPolicyConfig{
				"policy": &AnalyzerPolicyConfig{
					EnforcementLevel: "advisory",
					Properties: map[string]interface{}{
						"foo": float64(42),
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
			result, err := CreateConfigWithDefaults(test.Policies)
			assert.NoError(t, err)
			assert.Equal(t, test.Expected, result)
		})
	}
}

func TestValidatePolicyConfig(t *testing.T) {
	tests := []struct {
		Test     string
		Schema   AnalyzerPolicyConfigSchema
		Config   map[string]interface{}
		Expected []string
	}{
		{
			Test: "Required property missing",
			Schema: AnalyzerPolicyConfigSchema{
				Properties: map[string]JSONSchema{
					"foo": JSONSchema{
						"type": "string",
					},
				},
				Required: []string{"foo"},
			},
			Config:   map[string]interface{}{},
			Expected: []string{"foo is required"},
		},
		{
			Test: "Invalid type",
			Schema: AnalyzerPolicyConfigSchema{
				Properties: map[string]JSONSchema{
					"foo": JSONSchema{
						"type": "string",
					},
				},
			},
			Config: map[string]interface{}{
				"foo": 1,
			},
			Expected: []string{"foo: Invalid type. Expected: string, given: integer"},
		},
	}

	for _, test := range tests {
		t.Run(test.Test, func(t *testing.T) {
			result, err := validatePolicyConfig(test.Schema, test.Config)
			assert.NoError(t, err)
			assert.Equal(t, test.Expected, result)
		})
	}
}

// func TestReconcileSuccess(t *testing.T) {
// 	tests := []struct {
// 		Policies []AnalyzerPolicyInfo
// 		Config   map[string]*AnalyzerPolicyConfig
// 		Expected map[string]*AnalyzerPolicyConfig
// 	}{
// 		{
// 			Policies: []AnalyzerPolicyInfo{
// 				{
// 					Name:             "policy",
// 					EnforcementLevel: "advisory",
// 					ConfigSchema: &AnalyzerPolicyConfigSchema{
// 						Properties: map[string]JSONSchema{
// 							"foo": JSONSchema{
// 								"type":    "string",
// 								"default": "bar",
// 							},
// 						},
// 					},
// 				},
// 			},
// 			Config: map[string]*AnalyzerPolicyConfig{},
// 			Expected: map[string]*AnalyzerPolicyConfig{
// 				"policy": &AnalyzerPolicyConfig{
// 					EnforcementLevel: "advisory",
// 					Properties: map[string]interface{}{
// 						"foo": "bar",
// 					},
// 				},
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
// 			result, err := ReconcilePolicyPackConfig(test.Policies, test.Config)
// 			assert.NoError(t, err)
// 			assert.Equal(t, test.Expected, result)
// 		})
// 	}
// }

// func TestReconcileFail(t *testing.T) {
// 	tests := []struct {
// 		Policies []AnalyzerPolicyInfo
// 		Config   map[string]*AnalyzerPolicyConfig
// 		Expected error
// 	}{}

// 	for _, test := range tests {

// 	}
// }
