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

	"github.com/AlekSi/pointer"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// msgSetFreeMonitoring implements `setFreeMonitoring` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgSetFreeMonitoring(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	action, err := getRequiredParam[string](doc, "action")
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
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrBadValue,
			fmt.Sprintf(
				"Enumeration value '%s' for field '%s' is not a valid value.",
				action,
				doc.Command()+".action",
			),
			"action",
		)
	}

	if h.StateProvider.Get().TelemetryLocked {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrLocation50840,
			"Free Monitoring has been disabled via the command-line and/or config file",
			action,
		)
	}

	if err := h.StateProvider.Update(func(s *state.State) {
		if telemetryState {
			s.Telemetry = pointer.ToBool(true)
		} else {
			s.Telemetry = pointer.ToBool(false)
			s.LatestVersion = ""
			s.UpdateInfo = ""
			s.UpdateAvailable = false
		}
	}); err != nil {
		return nil, err
	}

	res := wirebson.MustDocument(
		"ok", float64(1),
	)

	return middleware.ResponseDoc(req, res)
}
