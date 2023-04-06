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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/blang/semver"
	"github.com/spf13/cobra"

	"github.com/pulumi/pulumi/sdk/v3/go/common/env"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

func newPluginRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "run KIND NAME[@VERSION] [ARGS]",
		Args:   cmdutil.MinimumNArgs(2),
		Hidden: !env.Dev.Value(),
		Short:  "Run a command on a plugin binary",
		Long: "Run a command on a plugin binary.\n" +
			"\n" +
			"Directly executes a plugin binary, if VERSION is not specified " +
			"the latest installed plugin will be used.",
		Run: cmdutil.RunFunc(func(cmd *cobra.Command, args []string) error {
			if !workspace.IsPluginKind(args[0]) {
				return fmt.Errorf("unrecognized plugin kind: %s", args[0])
			}
			kind := workspace.PluginKind(args[0])

			// Parse the name and version from the second argument in the form of "NAME[@VERSION]".
			name, version, found := strings.Cut(args[1], "@")
			if !tokens.IsName(name) {
				return fmt.Errorf("invalid plugin name %q", name)
			}
			var sv *semver.Version
			if found {
				v, err := semver.ParseTolerant(version)
				if err != nil {
					return fmt.Errorf("invalid plugin version %q: %w", version, err)
				}
				sv = &v
			}

			pluginDesc := fmt.Sprintf("%s %s", kind, name)
			if sv != nil {
				pluginDesc = fmt.Sprintf("%s@%s", pluginDesc, sv)
			}

			path, err := workspace.GetPluginPath(kind, name, sv, nil)
			if err != nil {
				return fmt.Errorf("could not get plugin path: %w", err)
			}

			pluginArgs := args[2:]

			pluginCmd := exec.Command(path, pluginArgs...)
			pluginCmd.Stdout = os.Stdout
			pluginCmd.Stderr = os.Stderr
			pluginCmd.Stdin = os.Stdin
			if err := pluginCmd.Start(); err != nil {
				if pathErr, ok := err.(*os.PathError); ok {
					syscallErr, ok := pathErr.Err.(syscall.Errno)
					if ok && syscallErr == syscall.ENOENT {
						return fmt.Errorf("could not find execute plugin %s, binary not found at %s", pluginDesc, path)
					}
				}
				return fmt.Errorf("could not execute plugin %s (%s): %w", pluginDesc, path, err)
			}
			err = pluginCmd.Wait()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						os.Exit(status.ExitStatus())
					}
				}
				return fmt.Errorf("could not execute plugin %s (%s): %w", pluginDesc, path, err)
			}

			return nil
		}),
	}

	return cmd
}
