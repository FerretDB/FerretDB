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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// SetFreeMonitoring is a part of common implementation of the setFreeMonitoring command.
func SetFreeMonitoring(_ context.Context, msg *wire.OpMsg, sp *state.Provider) (*wire.OpMsg, error) {
	if sp == nil {
		panic("state provider is required")
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()
	var action string
	if action, err = GetRequiredParam[string](document, "action"); err != nil {
		return nil, err
	}

	var telemetryState bool
	switch action {
	case "enable":
		telemetryState = true
	case "disable":
		telemetryState = false
	default:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf(
				"Enumeration value '%s' for field '%s' is not a valid value.",
				action,
				command+".action",
			),
			"action",
		)
	}

	if sp.Get().TelemetryLocked {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrFreeMonitoringDisabled,
			"Free Monitoring has been disabled via the command-line and/or config file",
			action,
		)
	}

	if err := sp.Update(func(s *state.State) {
		if telemetryState {
			s.EnableTelemetry()
		} else {
			s.DisableTelemetry()
		}
	}); err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
