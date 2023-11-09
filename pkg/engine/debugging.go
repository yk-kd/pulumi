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

package engine

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
)

func newDebuggingSink(events eventEmitter) plugin.DebuggingSink {
	return &debuggingSink{
		events: events,
	}
}

type debuggingSink struct {
	events eventEmitter // the channel to emit events into.
}

var _ plugin.DebuggingSink = (*debuggingSink)(nil)

func (s *debuggingSink) StartDebugging(info plugin.DebuggingInfo) error {
	s.events.startDebugging(info)
	return nil
}
