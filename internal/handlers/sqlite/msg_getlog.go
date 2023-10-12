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

package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetLog implements HandlerInterface.
func (h *Handler) MsgGetLog(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	getLog, err := document.Get(command)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, ok := getLog.(types.NullType); ok {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrMissingField,
			`BSON field 'getLog.getLog' is missing but a required field`,
		)
	}

	if _, ok := getLog.(string); !ok {
		return nil, commonerrors.NewCommandError(
			commonerrors.ErrTypeMismatch,
			fmt.Errorf(
				"BSON field 'getLog.getLog' is the wrong type '%s', expected type 'string'",
				commonparams.AliasFromType(getLog),
			),
		)
	}

	var resDoc *types.Document

	switch getLog {
	case "*":
		resDoc = must.NotFail(types.NewDocument(
			"names", must.NotFail(types.NewArray("global", "startupWarnings")),
			"ok", float64(1),
		))

	case "global":
		log, err := logging.RecentEntries.GetArray(zap.DebugLevel)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		resDoc = must.NotFail(types.NewDocument(
			"log", log,
			"totalLinesWritten", int64(log.Len()),
			"ok", float64(1),
		))

	case "startupWarnings":
		state := h.StateProvider.Get()

		info := version.Get()

		// it may be empty if no connection was established yet
		var b string
		if state.BackendVersion != "" {
			b, _, _ = strings.Cut(state.BackendVersion, " (")
			b = " and " + state.BackendName + " " + strings.TrimSpace(b)
		}

		startupWarnings := []string{
			fmt.Sprintf("Powered by FerretDB %s%s.", info.Version, b),
			"Please star us on GitHub: https://github.com/FerretDB/FerretDB.",
		}

		switch {
		case state.Telemetry == nil:
			startupWarnings = append(
				startupWarnings,
				"The telemetry state is undecided.",
				"Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.io.",
			)
		case state.UpdateAvailable:
			startupWarnings = append(
				startupWarnings,
				fmt.Sprintf(
					"A new version available! The latest version: %s. The current version: %s.",
					state.LatestVersion, info.Version,
				),
			)
		}

		var log types.Array

		for _, line := range startupWarnings {
			b, err := json.Marshal(map[string]any{
				"msg":  line,
				"tags": []string{"startupWarnings"},
				"s":    "I",
				"c":    "STORAGE",
				"id":   42000,
				"ctx":  "initandlisten",
				"t": map[string]string{
					"$date": time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00"),
				},
			})
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			log.Append(string(b))
		}
		resDoc = must.NotFail(types.NewDocument(
			"log", &log,
			"totalLinesWritten", int64(log.Len()),
			"ok", float64(1),
		))

	default:
		return nil, commonerrors.NewCommandError(
			commonerrors.ErrOperationFailed,
			fmt.Errorf("no RecentEntries named: %s", getLog),
		)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{resDoc},
	}))

	return &reply, nil
}
