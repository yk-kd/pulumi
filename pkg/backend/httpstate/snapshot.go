// Copyright 2016-2018, Pulumi Corporation.
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

package httpstate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mattbaird/jsonpatch"
	"github.com/pulumi/pulumi/pkg/v3/backend"
	"github.com/pulumi/pulumi/pkg/v3/backend/httpstate/client"
	"github.com/pulumi/pulumi/pkg/v3/resource/deploy"
	"github.com/pulumi/pulumi/pkg/v3/resource/stack"
	"github.com/pulumi/pulumi/pkg/v3/secrets"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

// cloudSnapshotPersister persists snapshots to the Pulumi service.
type cloudSnapshotPersister struct {
	context             context.Context         // The context to use for client requests.
	update              client.UpdateIdentifier // The UpdateIdentifier for this update sequence.
	tokenSource         *tokenSource            // A token source for interacting with the service.
	backend             *cloudBackend           // A backend for communicating with the service
	sm                  secrets.Manager
	sequence            int
	lastSavedDeployment *apitype.DeploymentV3
}

func (persister *cloudSnapshotPersister) SecretsManager() secrets.Manager {
	return persister.sm
}

func (persister *cloudSnapshotPersister) Save(snapshot *deploy.Snapshot) error {
	token, err := persister.tokenSource.GetToken()
	if err != nil {
		return err
	}
	deployment, err := stack.SerializeDeployment(snapshot, persister.sm, false /* showSecrets */)
	if err != nil {
		return fmt.Errorf("serializing deployment: %w", err)
	}
	previousDeployment := persister.lastSavedDeployment
	defer func() {
		persister.sequence++
		persister.lastSavedDeployment = deployment
	}()

	patch, err := serializeDeploymentAsJSONPatch(previousDeployment, deployment)
	if err != nil {
		return fmt.Errorf("serializing deployment patch: %w", err)
	}

	return persister.backend.client.PatchUpdateCheckpoint2(persister.context, persister.update, persister.sequence,
		patch, token, deployment)
}

var _ backend.SnapshotPersister = (*cloudSnapshotPersister)(nil)

func (cb *cloudBackend) newSnapshotPersister(ctx context.Context, update client.UpdateIdentifier,
	tokenSource *tokenSource, sm secrets.Manager) *cloudSnapshotPersister {
	return &cloudSnapshotPersister{
		context:     ctx,
		update:      update,
		tokenSource: tokenSource,
		backend:     cb,
		sm:          sm,
	}
}

func serializeDeploymentAsJSONPatch(previousDeployment, deployment *apitype.DeploymentV3) ([]byte, error) {
	var err error

	rawPreviousDeployment := []byte("{}")
	if previousDeployment != nil {
		rawPreviousDeployment, err = json.Marshal(previousDeployment)
		if err != nil {
			return nil, err
		}
	}

	rawDeployment, err := json.Marshal(deployment)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.CreatePatch(rawPreviousDeployment, rawDeployment)
	if err != nil {
		return nil, err
	}

	return json.Marshal(patch)
}
