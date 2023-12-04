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

package handler

import (
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgSetFreeMonitoring implements `setFreeMonitoring` command.
func (h *Handler) MsgSetFreeMonitoring(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	action, err := common.GetRequiredParam[string](document, "action")
	if err != nil {
		return nil, err
	}

	var telemetryState bool

	switch action {
	case "enable":
		telemetryState = true
	case "disable":
		telemetryState = false
	default:
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrBadValue,
			fmt.Sprintf(
				"Enumeration value '%s' for field '%s' is not a valid value.",
				action,
				document.Command()+".action",
			),
			"action",
		)
	}

	if h.StateProvider.Get().TelemetryLocked {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrFreeMonitoringDisabled,
			"Free Monitoring has been disabled via the command-line and/or config file",
			action,
		)
	}

	if err := h.StateProvider.Update(func(s *state.State) {
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
