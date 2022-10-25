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

package auto

import (
	"context"

	"github.com/pkg/errors"
)

// NewRemoteStackGitSource TODO doc comment
func NewRemoteStackGitSource(
	ctx context.Context,
	stackName string, repo GitRepo,
	opts ...RemoteWorkspaceOption,
) (RemoteStack, error) {
	localOpts, err := remoteToLocalOptions(repo, opts...)
	if err != nil {
		return RemoteStack{}, err
	}
	w, err := NewLocalWorkspace(ctx, localOpts...)
	if err != nil {
		return RemoteStack{}, errors.Wrap(err, "failed to create stack")
	}

	s, err := NewStack(ctx, stackName, w)
	if err != nil {
		return RemoteStack{}, err
	}
	return RemoteStack{stack: s}, nil
}

// UpsertRemoteStackGitSource TODO doc comment
func UpsertRemoteStackGitSource(
	ctx context.Context,
	stackName string, repo GitRepo,
	opts ...RemoteWorkspaceOption,
) (RemoteStack, error) {
	localOpts, err := remoteToLocalOptions(repo, opts...)
	if err != nil {
		return RemoteStack{}, err
	}
	w, err := NewLocalWorkspace(ctx, localOpts...)
	if err != nil {
		return RemoteStack{}, errors.Wrap(err, "failed to create stack")
	}

	s, err := NewStack(ctx, stackName, w)
	// error for all failures except if the stack already exists, as we'll
	// just select the stack if it exists.
	if err != nil && !IsCreateStack409Error(err) {
		return RemoteStack{stack: s}, err
	}

	return RemoteStack{stack: Stack{
		workspace: w,
		stackName: stackName,
	}}, nil
}

// SelectRemoteStackGitSource TODO doc comment
func SelectRemoteStackGitSource(
	ctx context.Context,
	stackName string, repo GitRepo,
	opts ...RemoteWorkspaceOption,
) (RemoteStack, error) {
	localOpts, err := remoteToLocalOptions(repo, opts...)
	if err != nil {
		return RemoteStack{}, err
	}
	w, err := NewLocalWorkspace(ctx, localOpts...)
	if err != nil {
		return RemoteStack{}, errors.Wrap(err, "failed to select stack")
	}

	return RemoteStack{stack: Stack{
		workspace: w,
		stackName: stackName,
	}}, nil
}

func remoteToLocalOptions(repo GitRepo, opts ...RemoteWorkspaceOption) ([]LocalWorkspaceOption, error) {
	if repo.Setup != nil {
		return nil, errors.New("repo.Setup cannot be used with remote workspaces")
	}

	remoteOpts := &remoteWorkspaceOptions{}
	for _, o := range opts {
		o.applyOption(remoteOpts)
	}

	localOpts := []LocalWorkspaceOption{
		remote(true),
		envVars(remoteOpts.EnvVars),
		preRunCommands(remoteOpts.PreRunCommands...),
		Repo(repo),
		SecretsProvider(remoteOpts.SecretsProvider),
	}
	return localOpts, nil
}

type remoteWorkspaceOptions struct {
	// SecretsProvider is the secrets provider to use with the remote
	// workspace when interacting with a stack.
	SecretsProvider string
	// EnvVars is a map of environment values scoped to the workspace.
	// These values will be passed to all Workspace and Stack level commands.
	EnvVars map[string]EnvVarValue
	// PreRunCommands is an optional list of arbitrary commands to run before the remote Pulumi operation is invoked.
	PreRunCommands []string
}

// LocalWorkspaceOption is used to customize and configure a LocalWorkspace at initialization time.
// See Workdir, Program, PulumiHome, Project, Stacks, and Repo for concrete options.
type RemoteWorkspaceOption interface {
	applyOption(*remoteWorkspaceOptions)
}

type remoteWorkspaceOption func(*remoteWorkspaceOptions)

func (o remoteWorkspaceOption) applyOption(opts *remoteWorkspaceOptions) {
	o(opts)
}

// RemoteSecretsProvider is the secrets provider to use with the remote
// workspace when interacting with a stack.
func RemoteSecretsProvider(secretsProvider string) RemoteWorkspaceOption {
	return remoteWorkspaceOption(func(opts *remoteWorkspaceOptions) {
		opts.SecretsProvider = secretsProvider
	})
}

// RemoteEnvVars is a map of environment values scoped to the remote workspace.
// These will be supplied to every Pulumi command and will be passed to remote operations.
func RemoteEnvVars(envvars map[string]EnvVarValue) RemoteWorkspaceOption {
	return remoteWorkspaceOption(func(opts *remoteWorkspaceOptions) {
		opts.EnvVars = envvars
	})
}

// RemotePreRunCommands is an optional list of arbitrary commands to run before the remote Pulumi operation is invoked.
func PreRunCommands(commands ...string) RemoteWorkspaceOption {
	return remoteWorkspaceOption(func(opts *remoteWorkspaceOptions) {
		opts.PreRunCommands = commands
	})
}
