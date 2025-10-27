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
	"log/slog"
	"strings"
	"time"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/devbuild"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// msgGetLog implements `getLog` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgGetLog(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	command := doc.Command()
	getLog := doc.Get(command)

	if _, ok := getLog.(wirebson.NullType); ok {
		return nil, mongoerrors.New(
			mongoerrors.ErrLocation40414,
			`BSON field 'getLog.getLog' is missing but a required field`,
		)
	}

	if _, ok := getLog.(string); !ok {
		return nil, mongoerrors.New(
			mongoerrors.ErrTypeMismatch,
			fmt.Sprintf(
				"BSON field 'getLog.getLog' is the wrong type '%s', expected type 'string'",
				aliasFromType(getLog),
			),
		)
	}

	var res *wirebson.Document

	switch getLog {
	case "*":
		res = wirebson.MustDocument(
			"names", wirebson.MustArray("global", "startupWarnings"),
			"ok", float64(1),
		)

	case "global":
		// TODO https://github.com/FerretDB/FerretDB/issues/4750
		log, err := h.L.Handler().(*logging.Handler).RecentEntries()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res = wirebson.MustDocument(
			"log", log,
			"totalLinesWritten", int32(log.Len()),
			"ok", float64(1),
		)

	case "startupWarnings":
		state := h.StateProvider.Get()

		info := version.Get()

		poweredBy := fmt.Sprintf("Powered by FerretDB %s", info.Version)

		// it may be empty if no connection was established yet
		if state.DocumentDBVersion != "" {
			v, _, _ := strings.Cut(state.DocumentDBVersion, " ")
			poweredBy += " and DocumentDB " + v + " ("

			v, _, _ = strings.Cut(state.PostgreSQLVersion, " (")
			poweredBy += v + ")"
		}

		poweredBy += "."

		startupWarnings := []string{
			poweredBy,
			"Please star 🌟 us on GitHub: https://github.com/FerretDB/FerretDB.",
		}

		if state.DocumentDBVersion != "" && state.DocumentDBVersion != version.DocumentDB {
			startupWarnings = append(
				startupWarnings,
				fmt.Sprintf(
					"This version of FerretDB requires DocumentDB %q (%s). The currently installed version is %q. "+
						"Some functions may not behave correctly.",
					version.DocumentDB, version.DocumentDBURL, state.DocumentDBVersion,
				),
			)
		}

		switch {
		case state.Telemetry == nil:
			startupWarnings = append(
				startupWarnings,
				"The telemetry state is undecided. "+
					"Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.com.",
			)

		case state.UpdateInfo != "", state.UpdateAvailable:
			msg := state.UpdateInfo
			if msg == "" {
				msg = fmt.Sprintf(
					"A new version is available! The latest version: %s. The current version: %s.",
					state.LatestVersion, info.Version,
				)
			}

			startupWarnings = append(startupWarnings, msg)
		}

		if devbuild.Enabled {
			startupWarnings = append(
				startupWarnings,
				"This is a development build. The performance will be affected.",
			)
		}

		if h.L.Enabled(connCtx, slog.LevelDebug) {
			startupWarnings = append(
				startupWarnings,
				"Debug logging enabled. The security and performance will be affected.",
			)
		}

		log := wirebson.MakeArray(len(startupWarnings))

		for _, line := range startupWarnings {
			ml := logging.MongoLogRecord{
				Msg:       line,
				Tags:      []string{"startupWarnings"},
				Severity:  "I",
				Component: "STORAGE",
				ID:        42000,
				Ctx:       "initandlisten",
				Timestamp: time.Now(),
			}

			var b []byte

			b, err := ml.Marshal()
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			if err = log.Add(string(b)); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
		res = wirebson.MustDocument(
			"log", log,
			"totalLinesWritten", int32(log.Len()),
			"ok", float64(1),
		)

	default:
		return nil, mongoerrors.New(
			mongoerrors.ErrOperationFailed,
			fmt.Sprintf("no RecentEntries named: %s", getLog),
		)
	}

	return middleware.ResponseDoc(req, res)
}
