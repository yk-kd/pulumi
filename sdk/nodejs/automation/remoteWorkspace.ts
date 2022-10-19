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

import { LocalWorkspace, LocalWorkspaceOptions } from "./localWorkspace";
import { RemoteStack } from "./remoteStack";
import { Stack } from "./stack";

/**
 * RemoteWorkspace is the execution context containing a single remote Pulumi project.
 */
export class RemoteWorkspace {
    /**
     * Creates a Stack with a LocalWorkspace utilizing the local Pulumi CLI program from the specified workDir.
     * This is a way to create drivers on top of pre-existing Pulumi programs. This Workspace will pick up
     * any available Settings files (Pulumi.yaml, Pulumi.<stack>.yaml).
     *
     * @param args A set of arguments to initialize a Stack with a pre-configured Pulumi CLI program that already exists on disk.
     * @param opts Additional customizations to be applied to the Workspace.
     */
    static async createStack(args: RemoteGitProgramArgs, opts?: RemoteWorkspaceOptions): Promise<RemoteStack> {
        const ws = await createLocalWorkspace(args, opts);
        const stack = await Stack.create(args.stackName, ws);
        return RemoteStack.create(stack);
    }

    /**
     * Selects a Stack with a LocalWorkspace utilizing the local Pulumi CLI program from the specified workDir.
     * This is a way to create drivers on top of pre-existing Pulumi programs. This Workspace will pick up
     * any available Settings files (Pulumi.yaml, Pulumi.<stack>.yaml).
     *
     * @param args A set of arguments to initialize a Stack with a pre-configured Pulumi CLI program that already exists on disk.
     * @param opts Additional customizations to be applied to the Workspace.
     */
    static async selectStack(args: RemoteGitProgramArgs, opts?: RemoteWorkspaceOptions): Promise<RemoteStack> {
        const ws = await createLocalWorkspace(args, opts);
        const stack = await Stack.selectWithOpts(args.stackName, ws, false /*select*/);
        return RemoteStack.create(stack);
    }
    /**
     * Creates or selects an existing Stack with a LocalWorkspace utilizing the specified inline (in process) Pulumi CLI program.
     * This program is fully debuggable and runs in process. If no Project option is specified, default project settings
     * will be created on behalf of the user. Similarly, unless a `workDir` option is specified, the working directory
     * will default to a new temporary directory provided by the OS.
     *
     * @param args A set of arguments to initialize a Stack with a pre-configured Pulumi CLI program that already exists on disk.
     * @param opts Additional customizations to be applied to the Workspace.
     */
    static async createOrSelectStack(args: RemoteGitProgramArgs, opts?: RemoteWorkspaceOptions): Promise<RemoteStack> {
        const ws = await createLocalWorkspace(args, opts);
        const stack = await Stack.createOrSelectWithOpts(args.stackName, ws, false /*select*/);
        return RemoteStack.create(stack);
    }

    private constructor() {} // eslint-disable-line @typescript-eslint/no-empty-function
}

/**
 * Description of a stack backed by a remote Pulumi program in a Git repository.
 */
export interface RemoteGitProgramArgs {
    /**
     * The name of the associated Stack
     */
    stackName: string;

    /**
     * The URL of the repository.
     */
    url: string;

    /**
     * Optional path relative to the repo root specifying location of the Pulumi program.
     */
    projectPath?: string;

    /**
     * Optional branch to checkout.
     */
    branch?: string;

    /**
     * Optional commit to checkout.
     */
    commitHash?: string;

    /**
     * Authentication options for the repository.
     */
    auth?: RemoteGitAuthArgs;
}

/**
 * Authentication options for the repository that can be specified for a private Git repo.
 * There are three different authentication paths:
 *  - Personal accesstoken
 *  - SSH private key (and its optional password)
 *  - Basic auth username and password
 *
 * Only one authentication path is valid.
 */
export interface RemoteGitAuthArgs {
    /**
     * The absolute path to a private key for access to the git repo.
     */
    sshPrivateKeyPath?: string;

    /**
     * The (contents) private key for access to the git repo.
     */
    sshPrivateKey?: string;

    /**
     * The password that pairs with a username or as part of an SSH Private Key.
     */
    password?: string;

    /**
     * PersonalAccessToken is a Git personal access token in replacement of your password.
     */
    personalAccessToken?: string;

    /**
     * Username is the username to use when authenticating to a git repository
     */
    username?: string;
}

/**
 * Extensibility options to configure a RemoteWorkspace.
 */
export interface RemoteWorkspaceOptions {
    /**
     * Environment values scoped to the remote workspace. These will be supplied to every Pulumi command and will be
     * passed to remote operations.
     */
    envVars?: { [key: string]: string | { secret: string } };

    /**
     * The secrets provider to use for encryption and decryption of stack secrets.
     * See: https://www.pulumi.com/docs/intro/concepts/secrets/#available-encryption-providers
     */
    secretsProvider?: string;

    /**
     * An optional list of arbitrary commands to run before the remote Pulumi operation is invoked.
     */
    preRunCommands?: string[];
}

function createLocalWorkspace(
    args: RemoteGitProgramArgs, opts?: RemoteWorkspaceOptions): Promise<LocalWorkspace> {
    const localOpts: LocalWorkspaceOptions = {
        ...opts,
        remote: true,
        remoteGitProgramArgs: args,
    };
    return LocalWorkspace.create(localOpts);
}
