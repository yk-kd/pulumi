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

package backend

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/pkg/v3/resource/deploy"
	"github.com/pulumi/pulumi/pkg/v3/resource/deploy/providers"
	"github.com/pulumi/pulumi/pkg/v3/resource/stack"
	"github.com/pulumi/pulumi/pkg/v3/secrets"
	"github.com/pulumi/pulumi/pkg/v3/version"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/config"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

// ISSUES:
//
// - default provider migration mutates the base snapshot in-place. particularly problematic because
//   it adds new resources to the state, which affects the indexing. it does, though, add all default
//   providers to the _front_ of the snapshot.
// - refreshes are issued out of dependency order (and actually affect the base state)
//
// The solution chosen here is to add a Rebase method to the SnapshotManager and JournalPersister
// interfaces that means "persist a new base state". This is be called after default provider migration
// and after a refresh finishes.

// JournalPersister is an interface implemented by our backends that implements journal
// persistence.
type JournalPersister interface {
	// Rebase persists a new base state for the journal. May only be called once and cannot be called
	// once Append has been called.
	Rebase(base *apitype.DeploymentV3) error

	// Persists the given entry. Returns an error if the persistence failed.
	Append(entry apitype.JournalEntry) error
}

var _ = engine.SnapshotManager((*Journal)(nil))

type aliasMap map[resource.URN]resource.URN

func (aliases aliasMap) normalizeURN(urn *resource.URN) {
	if u := *urn; u != "" {
		for {
			new, has := aliases[u]
			if !has {
				*urn = u
				return
			}
			u = new
		}
	}
}

func (aliases aliasMap) normalizeURNs(urns []resource.URN) {
	for i := range urns {
		aliases.normalizeURN(&urns[i])
	}
}

func (aliases aliasMap) normalizePropertyDependencies(m map[resource.PropertyKey][]resource.URN) {
	for _, urns := range m {
		aliases.normalizeURNs(urns)
	}
}

func (aliases aliasMap) normalizeProvider(provider *string) {
	if *provider != "" {
		ref, err := providers.ParseReference(*provider)
		contract.AssertNoErrorf(err, "malformed provider reference: %s", *provider)
		urn := ref.URN()
		aliases.normalizeURN(&urn)
		ref, err = providers.NewReference(urn, ref.ID())
		contract.AssertNoErrorf(err, "could not create provider reference with URN %s and ID %s", urn, ref.ID())
		*provider = ref.String()
	}
}

func (aliases aliasMap) normalize(r *apitype.ResourceV3) error {
	// TODO: what about resource references?

	for _, alias := range r.Aliases {
		if otherURN, has := aliases[alias]; has && otherURN != r.URN {
			return fmt.Errorf("Two resources ('%s' and '%s') aliases to the same: '%s'", otherURN, r.URN, alias)
		}
		aliases[alias] = r.URN
	}

	aliases.normalizeURN(&r.Parent)
	aliases.normalizeURNs(r.Dependencies)
	aliases.normalizePropertyDependencies(r.PropertyDependencies)
	aliases.normalizeProvider(&r.Provider)

	return nil
}

func verifyIntegrity(snap *apitype.DeploymentV3) error {
	if snap == nil {
		return nil
	}

	// Ensure the magic cookie checks out.
	if snap.Manifest.Magic != (deploy.Manifest{}).NewMagic() {
		return fmt.Errorf("magic cookie mismatch; possible tampering/corruption detected")
	}

	// Now check the resources.  For now, we just verify that parents come before children, and that there aren't
	// any duplicate URNs.
	urns := make(map[resource.URN]apitype.ResourceV3)
	provs := make(map[providers.Reference]struct{})
	for i, state := range snap.Resources {
		urn := state.URN

		if providers.IsProviderType(state.Type) {
			ref, err := providers.NewReference(urn, state.ID)
			if err != nil {
				return fmt.Errorf("provider %s is not referenceable: %v", urn, err)
			}
			provs[ref] = struct{}{}
		}
		if provider := state.Provider; provider != "" {
			ref, err := providers.ParseReference(provider)
			if err != nil {
				return fmt.Errorf("failed to parse provider reference for resource %s: %v", urn, err)
			}
			if _, has := provs[ref]; !has {
				return fmt.Errorf("resource %s refers to unknown provider %s", urn, ref)
			}
		}

		if par := state.Parent; par != "" {
			if _, has := urns[par]; !has {
				// The parent isn't there; to give a good error message, see whether it's missing entirely, or
				// whether it comes later in the snapshot (neither of which should ever happen).
				for _, other := range snap.Resources[i+1:] {
					if other.URN == par {
						return fmt.Errorf("child resource %s's parent %s comes after it", urn, par)
					}
				}
				return fmt.Errorf("child resource %s refers to missing parent %s", urn, par)
			}
		}

		for _, dep := range state.Dependencies {
			if _, has := urns[dep]; !has {
				// same as above - doing this for better error messages
				for _, other := range snap.Resources[i+1:] {
					if other.URN == dep {
						return fmt.Errorf("resource %s's dependency %s comes after it", urn, other.URN)
					}
				}

				return fmt.Errorf("resource %s dependency %s refers to missing resource", urn, dep)
			}
		}

		if _, has := urns[urn]; has && !state.Delete {
			// The only time we should have duplicate URNs is when all but one of them are marked for deletion.
			return fmt.Errorf("duplicate resource %s (not marked for deletion)", urn)
		}

		urns[urn] = state
	}

	return nil
}

type JournalReplayer struct {
	latest             int
	secrets            *apitype.SecretsProvidersV1
	pending            map[int]apitype.JournalEntry
	aliases            aliasMap
	resources          []apitype.ResourceV3
	news               map[int]int
	dones              map[int]bool
	pendingDeletion    map[int]bool
	pendingReplacement map[int]bool
}

func NewJournalReplayer() *JournalReplayer {
	return &JournalReplayer{
		pending:            map[int]apitype.JournalEntry{},
		aliases:            aliasMap{},
		news:               map[int]int{},
		dones:              map[int]bool{},
		pendingDeletion:    map[int]bool{},
		pendingReplacement: map[int]bool{},
	}
}

func (r *JournalReplayer) appendResource(res apitype.ResourceV3) {
	r.aliases.normalize(&res)
	r.resources = append(r.resources, res)
}

func (r *JournalReplayer) appendNewResource(id int, res apitype.ResourceV3) {
	r.news[id] = len(r.resources)
	r.appendResource(res)
}

func (r *JournalReplayer) Replay(e apitype.JournalEntry) error {
	if e.SequenceNumber != r.latest+1 {
		return fmt.Errorf("cannot replay entry %v out of order (latest is %v)", e.SequenceNumber, r.latest)
	}
	r.latest = e.SequenceNumber

	markDone := false
	switch e.Kind {
	case apitype.JournalEntryBegin:
		r.pending[e.SequenceNumber] = e
	case apitype.JournalEntryFailure:
		delete(r.pending, e.New)
	case apitype.JournalEntrySuccess:
		delete(r.pending, e.New)

		switch e.Op {
		case apitype.OpSame:
			r.appendNewResource(e.New, *e.State)
			markDone = e.Old != -1
		case apitype.OpUpdate, apitype.OpRefresh:
			r.appendNewResource(e.New, *e.State)
			markDone = true
		case apitype.OpCreate, apitype.OpCreateReplacement:
			r.appendNewResource(e.New, *e.State)
			markDone = r.pendingReplacement[e.Old]
		case apitype.OpDelete, apitype.OpDeleteReplaced, apitype.OpReadDiscard, apitype.OpDiscardReplaced:
			markDone = !r.pendingReplacement[e.Old]
		case apitype.OpReplace:
			// do nothing.
		case apitype.OpRead:
			r.appendNewResource(e.New, *e.State)
		case apitype.OpReadReplacement:
			r.appendNewResource(e.New, *e.State)
			markDone = true
		case apitype.OpRemovePendingReplace:
			markDone = true
		case apitype.OpImport, apitype.OpImportReplacement:
			r.appendNewResource(e.New, *e.State)
		}
	case apitype.JournalEntryPendingDeletion:
		r.pendingDeletion[e.Old] = true
	case apitype.JournalEntryPendingReplacement:
		r.pendingReplacement[e.Old] = true
	case apitype.JournalEntryOutputs:
		index, ok := r.news[e.New]
		if !ok {
			return fmt.Errorf("outputs entry refers to unknown resource %v", e.New)
		}
		r.resources[index].Outputs = e.State.Outputs
	case apitype.JournalEntrySecrets:
		if r.secrets != nil {
			return fmt.Errorf("secrets already recorded for this replay")
		}
		r.secrets = e.Secrets
	}

	if markDone {
		if e.Old == 0 {
			return fmt.Errorf("missing old ID for entry %v", e.SequenceNumber)
		}
		r.dones[e.Old] = true
	}
	return nil
}

func (r *JournalReplayer) Finish(base *apitype.DeploymentV3) (*apitype.DeploymentV3, error) {
	// Append any resources from the base snapshot that were not produced by the current snapshot.
	// See backend.SnapshotManager.snap for why this works.
	for i, res := range base.Resources {
		id := i + 1
		if !r.dones[id] {
			if r.pendingDeletion[id] {
				res.Delete = true
			}
			if r.pendingReplacement[id] {
				res.PendingReplacement = true
			}

			r.appendResource(res)
		}
	}

	// Append any pending operations.
	var ops []apitype.OperationV2
	if len(r.pending) != 0 {
		entries := make([]apitype.JournalEntry, 0, len(r.pending))
		for _, e := range r.pending {
			entries = append(entries, e)
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].SequenceNumber < entries[j].SequenceNumber })

		ops = make([]apitype.OperationV2, len(r.pending))
		for _, e := range entries {
			switch e.Op {
			case apitype.OpCreate, apitype.OpCreateReplacement:
				ops = append(ops, apitype.OperationV2{Resource: *e.State, Type: apitype.OperationTypeCreating})
			case apitype.OpDelete, apitype.OpDeleteReplaced, apitype.OpReadDiscard, apitype.OpDiscardReplaced:
				ops = append(ops, apitype.OperationV2{Resource: *e.State, Type: apitype.OperationTypeDeleting})
			case apitype.OpRead, apitype.OpReadReplacement:
				ops = append(ops, apitype.OperationV2{Resource: *e.State, Type: apitype.OperationTypeReading})
			case apitype.OpUpdate:
				ops = append(ops, apitype.OperationV2{Resource: *e.State, Type: apitype.OperationTypeUpdating})
			case apitype.OpImport, apitype.OpImportReplacement:
				ops = append(ops, apitype.OperationV2{Resource: *e.State, Type: apitype.OperationTypeImporting})
			}
		}
	}

	// Track pending create operations from the base snapshot
	// and propagate them to the new snapshot: we don't want to clear pending CREATE operations
	// because these must require user intervention to be cleared or resolved.
	for _, pendingOperation := range base.PendingOperations {
		if pendingOperation.Type == apitype.OperationTypeCreating {
			ops = append(ops, pendingOperation)
		}
	}

	d := &apitype.DeploymentV3{
		Manifest: apitype.ManifestV1{
			Time:    time.Now(),
			Magic:   base.Manifest.Magic,
			Version: version.Version,
		},
		SecretsProviders:  r.secrets,
		Resources:         r.resources,
		PendingOperations: ops,
	}
	return d, verifyIntegrity(d)
}

type appendRequest struct {
	kind  apitype.JournalEntryKind
	step  deploy.Step
	state *apitype.ResourceV3
}

type Journal struct {
	initOnce sync.Once
	initErr  error
	initDone bool

	hasRebase bool
	base      *deploy.Snapshot

	secrets   secrets.Manager
	enc       config.Encrypter
	olds      map[*resource.State]int
	news      sync.Map
	seq       atomic.Int32
	persister JournalPersister
}

func (j *Journal) markNew(s *resource.State, id int) {
	j.news.Store(s, id)
}

func (j *Journal) getNew(s *resource.State) int {
	if idAny, ok := j.news.Load(s); ok {
		return idAny.(int)
	}
	return 0
}

func (j *Journal) Close() error {
	return j.init()
}

func (j *Journal) Rebase(base *deploy.Snapshot) error {
	if j.initDone {
		return fmt.Errorf("Rebase may only be called before snapshot mutations begin")
	}
	j.base, j.hasRebase = base, true
	return nil
}

func (j *Journal) BeginMutation(step deploy.Step) (engine.SnapshotMutation, error) {
	if err := j.appendStep(apitype.JournalEntryBegin, step); err != nil {
		return nil, err
	}
	return j, nil
}

func (j *Journal) End(step deploy.Step, success bool) error {
	kind := apitype.JournalEntryFailure
	if success {
		kind = apitype.JournalEntrySuccess
	}
	return j.appendStep(kind, step)
}

func (j *Journal) RegisterResourceOutputs(step deploy.Step) error {
	return j.appendStep(apitype.JournalEntryOutputs, step)
}

func (j *Journal) append(entry apitype.JournalEntry) (int, error) {
	entry.SequenceNumber = int(j.seq.Add(1))
	if err := j.persister.Append(entry); err != nil {
		return 0, err
	}
	return entry.SequenceNumber, nil
}

func (j *Journal) getOld(kind apitype.JournalEntryKind, step deploy.Step) (int, error) {
	switch step.Op() {
	case deploy.OpImport, deploy.OpImportReplacement, deploy.OpRead:
		// Ignore olds for these ops.
		return 0, nil
	case deploy.OpSame:
		if step.(*deploy.SameStep).IsSkippedCreate() {
			// Ignore olds for skipped creates, but use a distinguished ID to indicate the skipped create.
			return -1, nil
		}
	}

	o := step.Old()
	if o == nil {
		return 0, nil
	}
	old, hasOld := j.olds[o]
	if !hasOld {
		return 0, fmt.Errorf("missing ID for old resource %v", step.URN())
	}

	if kind == apitype.JournalEntrySuccess {
		if o.Delete {
			if _, err := j.append(apitype.JournalEntry{Kind: apitype.JournalEntryPendingDeletion, Old: old}); err != nil {
				return 0, err
			}
		}
		if o.PendingReplacement {
			if _, err := j.append(apitype.JournalEntry{Kind: apitype.JournalEntryPendingReplacement, Old: old}); err != nil {
				return 0, err
			}
		}
	}

	return old, nil
}

func (j *Journal) appendStep(kind apitype.JournalEntryKind, step deploy.Step) error {
	// Refresh steps are not recorded. The outputs of these steps are expected to be captured by Rebase.
	if step.Op() == deploy.OpRefresh {
		return nil
	}

	if err := j.init(); err != nil {
		return err
	}

	var state *apitype.ResourceV3
	if r := step.Res(); r != nil {
		s, err := stack.SerializeResource(r, j.enc, false)
		if err != nil {
			return fmt.Errorf("serializing state: %w", err)
		}
		state = &s
	}

	old, err := j.getOld(kind, step)
	if err != nil {
		return err
	}

	var op apitype.OpType
	switch kind {
	case apitype.JournalEntryBegin, apitype.JournalEntryFailure, apitype.JournalEntrySuccess:
		op = apitype.OpType(step.Op())
	}

	seq, err := j.append(apitype.JournalEntry{
		Kind:  kind,
		Op:    op,
		Old:   old,
		New:   j.getNew(step.New()),
		State: state,
	})
	if err != nil {
		return err
	}
	if kind == apitype.JournalEntryBegin {
		j.markNew(step.New(), seq)
	}
	return nil
}

func (j *Journal) init() error {
	j.initOnce.Do(func() {
		j.initDone = true
		err := func() error {
			if j.hasRebase {
				base, err := stack.SerializeDeployment(j.base, nil, false)
				if err != nil {
					return fmt.Errorf("serializing deployment: %w", err)
				}
				if err = j.persister.Rebase(base); err != nil {
					return fmt.Errorf("rebasing: %w", err)
				}
			}

			olds := make(map[*resource.State]int, len(j.base.Resources))
			for i, r := range j.base.Resources {
				olds[r] = i + 1
			}
			j.olds = olds

			if j.secrets != nil {
				state, err := json.Marshal(j.secrets.State())
				if err != nil {
					return fmt.Errorf("serializing secret manager: %w", err)
				}
				_, err = j.append(apitype.JournalEntry{
					Kind: apitype.JournalEntrySecrets,
					Secrets: &apitype.SecretsProvidersV1{
						Type:  j.secrets.Type(),
						State: json.RawMessage(state),
					},
				})
				if err != nil {
					return fmt.Errorf("recording secret manager: %w", err)
				}
			}
			return nil
		}()
		if err != nil {
			j.initErr = err
		}
	})
	return j.initErr
}

func NewJournal(persister JournalPersister, base *deploy.Snapshot, sm secrets.Manager) (*Journal, error) {
	if sm == nil {
		sm = base.SecretsManager
	}

	var enc config.Encrypter
	if sm != nil {
		e, err := sm.Encrypter()
		if err != nil {
			return nil, fmt.Errorf("getting encrypter for deployment: %w", err)
		}
		enc = e
	} else {
		enc = config.NewPanicCrypter()
	}

	return &Journal{
		persister: persister,
		secrets:   sm,
		enc:       enc,
		base:      base,
	}, nil
}
