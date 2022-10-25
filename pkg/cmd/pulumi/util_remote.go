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
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pulumi/pulumi/pkg/v3/backend/display"
	"github.com/pulumi/pulumi/pkg/v3/backend/httpstate"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/result"
)

// This is a variable instead of a constant so it can be set in certain builds of the CLI that do not
// support remote deployments.
var disableRemote bool

// remoteSupported returns true if the CLI supports remote deployments.
func remoteSupported() bool {
	return !disableRemote && hasExperimentalCommands()
}

// parseEnv parses a `--remote-env` flag value for `--remote`. A value should be of the form
// "NAME=value" for regular envvars or "NAME#=value" for envvars with a secret value.
func parseEnv(input string) (string, string, bool, error) {
	pair := strings.SplitN(input, "=", 2)
	if len(pair) != 2 {
		return "", "", false, fmt.Errorf("error parsing %q", input)
	}
	name, value := pair[0], pair[1]

	var secret bool
	if last := len(name) - 1; last >= 0 && name[last] == '#' {
		secret = true
		name = name[:last]
	}

	if name == "" {
		return "", "", false, fmt.Errorf("expected non-empty environment name for %q", input)
	}

	return name, value, secret, nil
}

// validateUnsupportedRemoteFlags returns an error if any unsupported flags are set when --remote is set.
func validateUnsupportedRemoteFlags(
	expectNop bool,
	configArray []string,
	configPath bool,
	client string,
	jsonDisplay bool,
	policyPackPaths []string,
	policyPackConfigPaths []string,
	refresh string,
	showConfig bool,
	showReplacementSteps bool,
	showSames bool,
	showReads bool,
	suppressOutputs bool,
	secretsProvider string,
	targets *[]string,
	replaces []string,
	targetReplaces []string,
	targetDependents bool,
	planFilePath string,
	stackConfigFile string,
) error {

	if expectNop {
		return errors.New("--expect-no-changes is not supported with --remote")
	}
	if len(configArray) > 0 {
		return errors.New("--config is not supported with --remote")
	}
	if configPath {
		return errors.New("--config-path is not supported with --remote")
	}
	if client != "" {
		return errors.New("--client is not supported with --remote")
	}
	// We should be able to make --json work, but it doesn't work currently.
	if jsonDisplay {
		return errors.New("--json is not supported with --remote")
	}
	if len(policyPackPaths) > 0 {
		return errors.New("--policy-pack is not supported with --remote")
	}
	if len(policyPackConfigPaths) > 0 {
		return errors.New("--policy-pack-config is not supported with --remote")
	}
	if refresh != "" {
		return errors.New("--refresh is not supported with --remote")
	}
	if showConfig {
		return errors.New("--show-config is not supported with --remote")
	}
	if showReplacementSteps {
		return errors.New("--show-replacement-steps is not supported with --remote")
	}
	if showSames {
		return errors.New("--show-sames is not supported with --remote")
	}
	if showReads {
		return errors.New("--show-reads is not supported with --remote")
	}
	if suppressOutputs {
		return errors.New("--suppress-outputs is not supported with --remote")
	}
	if secretsProvider != "default" {
		return errors.New("--secrets-provider is not supported with --remote")
	}
	if targets != nil && len(*targets) > 0 {
		return errors.New("--target is not supported with --remote")
	}
	if len(replaces) > 0 {
		return errors.New("--replace is not supported with --remote")
	}
	if len(replaces) > 0 {
		return errors.New("--replace is not supported with --remote")
	}
	if len(targetReplaces) > 0 {
		return errors.New("--target-replace is not supported with --remote")
	}
	if targetDependents {
		return errors.New("--target-dependents is not supported with --remote")
	}
	if planFilePath != "" {
		return errors.New("--plan is not supported with --remote")
	}
	if stackConfigFile != "" {
		return errors.New("--config-file is not supported with --remote")
	}

	return nil
}

// runDeployment kicks off a remote deployment.
func runDeployment(ctx context.Context, opts display.Options, operation apitype.PulumiOperation, stack string,
	envVars, preRunCommands []string, gitRepoURL, gitBranch, gitCommit, gitRepoDir,
	gitAuthAccessToken, gitAuthSSHPrivateKey, gitAuthSSHPrivateKeyPath, gitAuthPassword,
	gitAuthUsername string) result.Result {

	b, err := currentBackend(ctx, opts)
	if err != nil {
		return result.FromError(err)
	}

	// Ensure the cloud backend is being used.
	cb, isCloud := b.(httpstate.Backend)
	if !isCloud {
		return result.FromError(errors.New("the Pulumi service backend must be used for remote operations; " +
			"use `pulumi login` without arguments to log into the Pulumi service backend"))
	}

	stackRef, err := b.ParseStackReference(stack)
	if err != nil {
		return result.FromError(err)
	}

	if gitRepoURL == "" {
		return result.FromError(errors.New("Git repo URL most be specified"))
	}
	if gitCommit != "" && gitBranch != "" {
		return result.FromError(errors.New("Git commit and branch cannot both be specified"))
	}
	if gitCommit == "" && gitBranch == "" {
		return result.FromError(errors.New("at least Git commit or branch are required"))
	}

	env := map[string]apitype.SecretValue{}
	for _, e := range envVars {
		name, value, secret, err := parseEnv(e)
		if err != nil {
			return result.FromError(err)
		}
		env[name] = apitype.SecretValue{Value: value, Secret: secret}
	}

	var gitAuth *apitype.GitAuthConfig
	if gitAuthAccessToken != "" || gitAuthSSHPrivateKey != "" || gitAuthSSHPrivateKeyPath != "" ||
		gitAuthPassword != "" || gitAuthUsername != "" {

		gitAuth = &apitype.GitAuthConfig{}
		switch {
		case gitAuthAccessToken != "":
			gitAuth.PersonalAccessToken = &apitype.SecretValue{Value: gitAuthAccessToken, Secret: true}

		case gitAuthSSHPrivateKey != "" || gitAuthSSHPrivateKeyPath != "":
			sshAuth := &apitype.SSHAuth{}
			if gitAuthSSHPrivateKeyPath != "" {
				content, err := os.ReadFile(gitAuthSSHPrivateKeyPath)
				if err != nil {
					return result.FromError(fmt.Errorf(
						"reading SSH private key path %q: %w", gitAuthSSHPrivateKeyPath, err))
				}
				sshAuth.SSHPrivateKey = apitype.SecretValue{Value: string(content), Secret: true}
			} else {
				sshAuth.SSHPrivateKey = apitype.SecretValue{Value: gitAuthSSHPrivateKey, Secret: true}
			}
			if gitAuthPassword != "" {
				sshAuth.Password = &apitype.SecretValue{Value: gitAuthPassword, Secret: true}
			}
			gitAuth.SSHAuth = sshAuth

		case gitAuthUsername != "":
			basicAuth := &apitype.BasicAuth{UserName: apitype.SecretValue{Value: gitAuthUsername, Secret: true}}
			if gitAuthPassword != "" {
				basicAuth.Password = apitype.SecretValue{Value: gitAuthPassword, Secret: true}
			}
			gitAuth.BasicAuth = basicAuth
		}
	}

	req := apitype.CreateDeploymentRequest{
		Source: &apitype.SourceContext{
			Git: &apitype.SourceContextGit{
				RepoURL: gitRepoURL,
				Branch:  gitBranch,
				RepoDir: gitRepoDir,
				GitAuth: gitAuth,
			},
		},
		Operation: &apitype.OperationContext{
			Operation:            operation,
			PreRunCommands:       preRunCommands,
			EnvironmentVariables: env,
		},
	}
	err = cb.RunDeployment(ctx, stackRef, req, opts)
	if err != nil {
		return result.FromError(err)
	}

	return nil
}
