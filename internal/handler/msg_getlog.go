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
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/build/version"
	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgGetLog implements `getLog` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgGetLog(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := bson.OpMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	getLog, err := document.Get(command)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, ok := getLog.(types.NullType); ok {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrMissingField,
			`BSON field 'getLog.getLog' is missing but a required field`,
		)
	}

	if _, ok := getLog.(string); !ok {
		return nil, handlererrors.NewCommandError(
			handlererrors.ErrTypeMismatch,
			fmt.Errorf(
				"BSON field 'getLog.getLog' is the wrong type '%s', expected type 'string'",
				handlerparams.AliasFromType(getLog),
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
		var res *wirebson.Array

		if res, err = logging.RecentEntries.GetArray(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		resDoc = must.NotFail(types.NewDocument(
			"log", must.NotFail(bson.TypesArray(res)),
			"totalLinesWritten", int64(res.Len()),
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

		if debugbuild.Enabled {
			startupWarnings = append(startupWarnings, "This is debug build. The performance will be affected.")
		}

		if h.L.Enabled(connCtx, slog.LevelDebug) {
			startupWarnings = append(startupWarnings, "Debug logging enabled. The security and performance will be affected.")
		}

		switch {
		case state.Telemetry == nil:
			startupWarnings = append(
				startupWarnings,
				"The telemetry state is undecided.",
				"Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.com.",
			)

		case state.UpdateInfo != "", state.UpdateAvailable:
			msg := state.UpdateInfo
			if msg == "" {
				msg = fmt.Sprintf(
					"A new version available! The latest version: %s. The current version: %s.",
					state.LatestVersion, info.Version,
				)
			}

			startupWarnings = append(startupWarnings, msg)
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
		return nil, handlererrors.NewCommandError(
			handlererrors.ErrOperationFailed,
			fmt.Errorf("no RecentEntries named: %s", getLog),
		)
	}

	return bson.NewOpMsg(
		resDoc,
	)
}
