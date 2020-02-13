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
	"io/ioutil"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/apitype"
)

// LoadPolicyPackConfigFromFile loads the JSON config from a file.
func LoadPolicyPackConfigFromFile(file string) (map[string]AnalyzerPolicyConfig, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return parsePolicyPackConfig(b)
}

func parsePolicyPackConfig(b []byte) (map[string]AnalyzerPolicyConfig, error) {
	config := make(map[string]interface{})
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	// TODO
	return make(map[string]AnalyzerPolicyConfig), nil
}

// CheckRequired checks the given config to ensure all required properties are set.
func CheckRequired(policies []AnalyzerPolicyInfo, config map[string]*AnalyzerPolicyConfig) error {
	var result error

	// TODO provide a more elegant way of returning multiple errors, so we can display all of these
	// up-front.
	// TODO would it be easier to use a synthesized schema for the policy, and just let the validator
	// check this?
	for _, policy := range policies {
		// If the policy doesn't have config schema, skip.
		if policy.Config == nil {
			continue
		}
		for _, required := range policy.Config.Required {
			givenConfig, hasGivenConfig := config[policy.Name]
			if !hasGivenConfig {
				result = multierror.Append(
					result, errors.Errorf("Missing required property %q for policy %q", required, policy.Name))
			}
			if _, hasProperty := givenConfig.Properties[required]; !hasProperty {
				result = multierror.Append(
					errors.Errorf("Missing required property %q for policy %q", required, policy.Name))
			}
		}
	}

	return result
}

// CreateConfigWithDefaults returns a new map filled-in with defaults from the policy metadata.
func CreateConfigWithDefaults(policies []AnalyzerPolicyInfo) (map[string]*AnalyzerPolicyConfig, error) {
	result := make(map[string]*AnalyzerPolicyConfig)

	// Prepare the resulting config with all defaults from the policy metadata.
	for _, policy := range policies {
		var props map[string]interface{}

		// Set default values from the schema.
		if policy.Config != nil {
			props = make(map[string]interface{})
			for k, v := range policy.Config.Properties {
				schema, err := unmarshalConfigSchema([]byte(v))
				if err != nil {
					return nil, err
				}
				if val, ok := schema["default"]; ok {
					props[k] = val
				}
			}
		}

		result[policy.Name] = &AnalyzerPolicyConfig{
			EnforcementLevel: policy.EnforcementLevel,
			Properties:       props,
		}
	}

	return result, nil
}

// ReconcilePolicyPackConfig takes the policy pack metadata, which contains default values and config schema,
// and reconciles this with the given config, producing a new config with all default values filled-in and then
// any given config values filled-in on top, to be passed to the analyzer.
func ReconcilePolicyPackConfig(
	policies []AnalyzerPolicyInfo, config map[string]*AnalyzerPolicyConfig) (map[string]*AnalyzerPolicyConfig, error) {

	// First, loop through the given config, and ensure we have values for all required properties.
	if err := CheckRequired(policies, config); err != nil {
		return nil, err
	}

	// Next, prepare the resulting config with all defaults from the policy metadata.
	result, err := CreateConfigWithDefaults(policies)
	if err != nil {
		return nil, err
	}

	// Next, if the given config has "all" and an enforcement level, set it for all
	// policies.
	if all, hasAll := config["all"]; hasAll && all.EnforcementLevel != "" {
		for _, v := range result {
			v.EnforcementLevel = all.EnforcementLevel
		}
	}

	// TODO validate the given config with the schema.

	// Next, loop through the given config, and set values.
	for policy, givenConfig := range config {
		// TODO should we error or warn if config has a policy name that isn't in policies?
		resultConfig, hasResultConfig := result[policy]
		if !hasResultConfig {
			continue
		}

		if givenConfig.EnforcementLevel != "" {
			resultConfig.EnforcementLevel = givenConfig.EnforcementLevel
		}
		if len(givenConfig.Properties) > 0 && resultConfig.Properties == nil {
			resultConfig.Properties = make(map[string]interface{})
		}
		for k, v := range givenConfig.Properties {
			resultConfig.Properties[k] = v
		}
	}

	return result, nil
}

func unmarshalConfigSchema(b []byte) (map[string]interface{}, error) {
	schema := make(map[string]interface{})
	if err := json.Unmarshal(b, &schema); err != nil {
		return nil, err
	}
	return schema, nil
}

type policyConfig struct {
	// Configured enforcement level for the policy.
	EnforcementLevel apitype.EnforcementLevel
	// Configuration properties of the policy.
	Properties map[string]interface{}
}
