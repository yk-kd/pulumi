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
	"fmt"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto/events"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optremotepreview"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
)

func TestIsFullyQualifiedStackName(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    string
		expected bool
	}{
		"fully qualified": {input: "owner/project/stack", expected: true},
		"empty":           {input: "", expected: false},
		"name":            {input: "name", expected: false},
		"name & owner":    {input: "owner/name", expected: false},
		"sep":             {input: "/", expected: false},
		"two seps":        {input: "//", expected: false},
		"three seps":      {input: "///", expected: false},
		"invalid":         {input: "owner/project/stack/wat", expected: false},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := isFullyQualifiedStackName(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

const remoteTestRepo = "https://github.com/pulumi/test-repo.git"

func testRemoteStackGitSourceErrors(t *testing.T, fn func(ctx context.Context, stackName string, repo GitRepo,
	opts ...RemoteWorkspaceOption) (RemoteStack, error)) {

	ctx := context.Background()

	const stack = "owner/project/stack"

	tests := map[string]struct {
		stack string
		repo  GitRepo
		err   string
	}{
		"stack empty": {
			stack: "",
			err:   `"" stack name must be fully qualified`,
		},
		"stack just name": {
			stack: "name",
			err:   `"name" stack name must be fully qualified`,
		},
		"stack just name & owner": {
			stack: "owner/name",
			err:   `"owner/name" stack name must be fully qualified`,
		},
		"stack just sep": {
			stack: "/",
			err:   `"/" stack name must be fully qualified`,
		},
		"stack just two seps": {
			stack: "//",
			err:   `"//" stack name must be fully qualified`,
		},
		"stack just three seps": {
			stack: "///",
			err:   `"///" stack name must be fully qualified`,
		},
		"stack invalid": {
			stack: "owner/project/stack/wat",
			err:   `"owner/project/stack/wat" stack name must be fully qualified`,
		},
		"repo setup": {
			stack: stack,
			repo:  GitRepo{Setup: func(context.Context, Workspace) error { return nil }},
			err:   "repo.Setup cannot be used with remote workspaces",
		},
		"no url": {
			stack: stack,
			repo:  GitRepo{},
			err:   "repo.URL is required",
		},
		"no branch or commit": {
			stack: stack,
			repo:  GitRepo{URL: remoteTestRepo},
			err:   "at least repo.CommitHash or repo.Branch are required",
		},
		"both branch and commit": {
			stack: stack,
			repo:  GitRepo{URL: remoteTestRepo, Branch: "branch", CommitHash: "commit"},
			err:   "repo.CommitHash and repo.Branch cannot both be specified",
		},
		"both ssh private key and path": {
			stack: stack,
			repo: GitRepo{
				URL:    remoteTestRepo,
				Branch: "branch",
				Auth:   &GitAuth{SSHPrivateKey: "key", SSHPrivateKeyPath: "path"},
			},
			err: "repo.Auth.SSHPrivateKey and repo.Auth.SSHPrivateKeyPath cannot both be specified",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := fn(ctx, tc.stack, tc.repo)
			assert.EqualError(t, err, tc.err)
		})
	}
}

func testRemoteStackGitSource(t *testing.T, fn func(ctx context.Context, stackName string, repo GitRepo,
	opts ...RemoteWorkspaceOption) (RemoteStack, error)) {

	ctx := context.Background()
	pName := "go_remote_proj"
	sName := randomStackName()
	stackName := FullyQualifiedStackName(pulumiOrg, pName, sName)
	repo := GitRepo{
		URL:         remoteTestRepo,
		Branch:      "refs/heads/master",
		ProjectPath: "goproj",
	}

	// initialize
	s, err := fn(ctx, stackName, repo, RemotePreRunCommands(
		fmt.Sprintf("pulumi config set bar abc --stack %s", stackName),
		fmt.Sprintf("pulumi config set --secret buzz secret --stack %s", stackName)))
	if err != nil {
		t.Errorf("failed to initialize stack, err: %v", err)
		t.FailNow()
	}

	defer func() {
		// -- pulumi stack rm --
		err = s.stack.Workspace().RemoveStack(ctx, s.Name())
		assert.Nil(t, err, "failed to remove stack. Resources have leaked.")
	}()

	// -- pulumi up --
	res, err := s.Up(ctx)
	if err != nil {
		t.Errorf("up failed, err: %v", err)
		t.FailNow()
	}

	assert.Equal(t, 3, len(res.Outputs), "expected two plain outputs")
	assert.Equal(t, "foo", res.Outputs["exp_static"].Value)
	assert.False(t, res.Outputs["exp_static"].Secret)
	assert.Equal(t, "abc", res.Outputs["exp_cfg"].Value)
	assert.False(t, res.Outputs["exp_cfg"].Secret)
	assert.Equal(t, "secret", res.Outputs["exp_secret"].Value)
	assert.True(t, res.Outputs["exp_secret"].Secret)
	assert.Equal(t, "update", res.Summary.Kind)
	assert.Equal(t, "succeeded", res.Summary.Result)

	// -- pulumi preview --

	var previewEvents []events.EngineEvent
	prevCh := make(chan events.EngineEvent)
	wg := collectEvents(prevCh, &previewEvents)
	prev, err := s.Preview(ctx, optremotepreview.EventStreams(prevCh))
	if err != nil {
		t.Errorf("preview failed, err: %v", err)
		t.FailNow()
	}
	wg.Wait()
	assert.Equal(t, 1, prev.ChangeSummary[apitype.OpSame])
	steps := countSteps(previewEvents)
	assert.Equal(t, 1, steps)

	// -- pulumi refresh --

	ref, err := s.Refresh(ctx)

	if err != nil {
		t.Errorf("refresh failed, err: %v", err)
		t.FailNow()
	}
	assert.Equal(t, "refresh", ref.Summary.Kind)
	assert.Equal(t, "succeeded", ref.Summary.Result)

	// -- pulumi destroy --

	dRes, err := s.Destroy(ctx)
	if err != nil {
		t.Errorf("destroy failed, err: %v", err)
		t.FailNow()
	}

	assert.Equal(t, "destroy", dRes.Summary.Kind)
	assert.Equal(t, "succeeded", dRes.Summary.Result)
}

func TestSelectRemoteStackGitSourceErrors(t *testing.T) {
	t.Parallel()
	testRemoteStackGitSourceErrors(t, SelectRemoteStackGitSource)
}

func TestNewRemoteStackGitSourceErrors(t *testing.T) {
	t.Parallel()
	testRemoteStackGitSourceErrors(t, NewRemoteStackGitSource)
}

func TestNewRemoteStackGitSource(t *testing.T) {
	t.Parallel()
	testRemoteStackGitSource(t, NewRemoteStackGitSource)
}

func TestUpsertRemoteStackGitSourceErrors(t *testing.T) {
	t.Parallel()
	testRemoteStackGitSourceErrors(t, UpsertRemoteStackGitSource)
}

func TestUpsertRemoteStackGitSource(t *testing.T) {
	t.Parallel()
	testRemoteStackGitSource(t, UpsertRemoteStackGitSource)
}
