# Copyright 2016-2022, Pulumi Corporation.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from typing import List, Mapping, Optional, Union

from pulumi.automation._local_workspace import LocalWorkspace
from pulumi.automation._remote_stack import RemoteStack
from pulumi.automation._stack import Stack, StackInitMode


class Secret:
    def __init__(self, secret: str):
        self.secret = secret


class RemoteWorkspaceOptions:
    env_vars: Optional[Mapping[str, Union[str, Secret]]]
    secrets_provider: Optional[str]
    pre_run_commands: Optional[List[str]]

    def __init__(
        self,
        env_vars: Optional[Mapping[str, Union[str, Secret]]] = None,
        secrets_provider: Optional[str] = None,
        pre_run_commands: Optional[List[str]] = None,
    ):
        self.env_vars = env_vars
        self.secrets_provider = secrets_provider
        self.pre_run_commands = pre_run_commands


class RemoteGitAuth:
    ssh_private_key_path: Optional[str]
    """
    The absolute path to a private key for access to the git repo.
    """

    ssh_private_key: Optional[str]
    """
    The (contents) private key for access to the git repo.
    """

    password: Optional[str]
    """
    The password that pairs with a username or as part of an SSH Private Key.
    """

    personal_access_token: Optional[str]
    """
    A Git personal access token in replacement of your password.
    """

    username: Optional[str]
    """
    The username to use when authenticating to a git repository.
    """

    def __init__(
        self,
        *,
        ssh_private_key_path: Optional[str] = None,
        ssh_private_key: Optional[str] = None,
        password: Optional[str] = None,
        personal_access_token: Optional[str] = None,
        username: Optional[str] = None,
    ):
        self.ssh_private_key_path = ssh_private_key_path
        self.ssh_private_key = ssh_private_key
        self.password = password
        self.personal_access_token = personal_access_token
        self.username = username


def create_remote_stack_git_source(
    stack_name: str,
    url: str,
    project_path: Optional[str] = None,
    branch: Optional[str] = None,
    commit_hash: Optional[str] = None,
    auth: Optional[RemoteGitAuth] = None,
    opts: Optional[RemoteWorkspaceOptions] = None,
) -> RemoteStack:
    ws = _create_local_workspace(
        url=url,
        project_path=project_path,
        branch=branch,
        commit_hash=commit_hash,
        auth=auth,
        opts=opts,
    )
    stack = Stack.create(stack_name, ws)
    return RemoteStack(stack)


def create_or_select_remote_stack_git_source(
    stack_name: str,
    url: str,
    project_path: Optional[str] = None,
    branch: Optional[str] = None,
    commit_hash: Optional[str] = None,
    auth: Optional[RemoteGitAuth] = None,
    opts: Optional[RemoteWorkspaceOptions] = None,
) -> RemoteStack:
    ws = _create_local_workspace(
        url=url,
        project_path=project_path,
        branch=branch,
        commit_hash=commit_hash,
        auth=auth,
        opts=opts,
    )
    stack = Stack(stack_name, ws, StackInitMode.CREATE_OR_SELECT, select=False)
    return RemoteStack(stack)


def select_remote_stack_git_source(
    stack_name: str,
    url: str,
    project_path: Optional[str] = None,
    branch: Optional[str] = None,
    commit_hash: Optional[str] = None,
    auth: Optional[RemoteGitAuth] = None,
    opts: Optional[RemoteWorkspaceOptions] = None,
) -> RemoteStack:
    ws = _create_local_workspace(
        url=url,
        project_path=project_path,
        branch=branch,
        commit_hash=commit_hash,
        auth=auth,
        opts=opts,
    )
    stack = Stack(stack_name, ws, StackInitMode.SELECT, select=False)
    return RemoteStack(stack)


def _create_local_workspace(
    url: str,
    project_path: Optional[str] = None,
    branch: Optional[str] = None,
    commit_hash: Optional[str] = None,
    auth: Optional[RemoteGitAuth] = None,
    opts: Optional[RemoteWorkspaceOptions] = None,
) -> LocalWorkspace:
    secrets_provider = None
    env_vars = None
    pre_run_commands = None
    if opts is not None:
        secrets_provider = opts.secrets_provider
        env_vars = opts.env_vars
        pre_run_commands = opts.pre_run_commands

    ws = LocalWorkspace(secrets_provider=secrets_provider)
    ws._remote = True
    ws._remote_env_vars = env_vars
    ws._remote_pre_run_commands = pre_run_commands
    ws._remote_git_url = url
    ws._remote_git_project_path = project_path
    ws._remote_git_branch = branch
    ws._remote_git_commit_hash = commit_hash
    ws._remote_git_auth = auth
    return ws
