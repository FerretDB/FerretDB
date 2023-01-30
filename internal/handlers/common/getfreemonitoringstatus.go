// Copyright 2021 FerretDB Inc.
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

package common

import (
	"context"

	"github.com/AlekSi/pointer"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// GetFreeMonitoringStatus is a part of common implementation of the getFreeMonitoringStatus command.
func GetFreeMonitoringStatus(ctx context.Context, msg *wire.OpMsg, state *state.State) (*wire.OpMsg, error) {
	if state == nil {
		panic("state cannot be equal to nil")
	}

	telemetryState := "disabled"
	telemetryMsg := "monitoring is not enabled"

	switch {
	case state.Telemetry == nil:
		telemetryState = "undecided"
		telemetryMsg = "monitoring is undecided"
	case pointer.GetBool(state.Telemetry):
		telemetryState = "enabled"
		telemetryMsg = "monitoring is enabled"
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"state", telemetryState,
			"message", telemetryMsg,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
