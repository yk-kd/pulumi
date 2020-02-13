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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
					Config: &AnalyzerPolicyConfigInfo{
						Properties: map[string]string{
							"foo": asJSON(t, map[string]interface{}{
								"type":    "string",
								"default": "bar",
							}),
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
					Config: &AnalyzerPolicyConfigInfo{
						Properties: map[string]string{
							"foo": asJSON(t, map[string]interface{}{
								"type":    "number",
								"default": 42,
							}),
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

func TestReconcileSuccess(t *testing.T) {
	tests := []struct {
		Policies []AnalyzerPolicyInfo
		Config   map[string]*AnalyzerPolicyConfig
		Expected map[string]*AnalyzerPolicyConfig
	}{
		{
			Policies: []AnalyzerPolicyInfo{
				{
					Name:             "policy",
					EnforcementLevel: "advisory",
					Config: &AnalyzerPolicyConfigInfo{
						Properties: map[string]string{
							"foo": asJSON(t, map[string]interface{}{
								"type":    "string",
								"default": "bar",
							}),
						},
					},
				},
			},
			Config: map[string]*AnalyzerPolicyConfig{},
			Expected: map[string]*AnalyzerPolicyConfig{
				"policy": &AnalyzerPolicyConfig{
					EnforcementLevel: "advisory",
					Properties: map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test), func(t *testing.T) {
			result, err := ReconcilePolicyPackConfig(test.Policies, test.Config)
			assert.NoError(t, err)
			assert.Equal(t, test.Expected, result)
		})
	}
}

// func TestReconcileFail(t *testing.T) {
// 	tests := []struct {
// 		Policies []AnalyzerPolicyInfo
// 		Config   map[string]*AnalyzerPolicyConfig
// 		Expected error
// 	}{}

// 	for _, test := range tests {

// 	}
// }

func asJSON(t *testing.T, schema map[string]interface{}) string {
	b, err := json.Marshal(schema)
	assert.NoError(t, err)
	return string(b)
}
