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

package pg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
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
	if _, ok := getLog.(string); !ok {
		// 	"errmsg" : "Argument to getLog must be of type String; found nan.0 of type double",
		// 	"code" : 14.0,
		// 	"codeName" : "TypeMismatch"
	}

	resDoc := must.NotFail(types.NewDocument())

	switch getLog {
	case "*":
		resDoc = must.NotFail(types.NewDocument(
			"names", must.NotFail(types.NewArray("global", "startupWarnings")),
			"ok", float64(1),
		))

	case "global":
		entries := logging.RecentEntries.Get()
		var log types.Array
		for _, e := range entries {
			b, err := json.Marshal(map[string]any{
				"t":   e.Time,
				"l":   e.Level,
				"ln":  e.LoggerName,
				"msg": e.Message,
				"c":   e.Caller,
				"s":   e.Stack,
			})
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if err = log.Append(string(b)); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
		resDoc = must.NotFail(types.NewDocument(
			"log", &log,
			"totalLinesWritten", len(entries),
			"ok", float64(1),
		))

	case "startupWarnings":

	default:
		errMsg := fmt.Sprintf("no RamLog named: %s", getLog)
		return nil, common.NewErrorMsg(0, errMsg)
	}

	// if l := document.Map()["getLog"]; l != "startupWarnings" {
	// 	errMsg := fmt.Sprintf("MsgGetLog: unhandled getLog value %q", l)
	// 	return nil, common.NewErrorMsg(common.ErrNotImplemented, errMsg)
	// }

	// var pv string
	// err = h.pgPool.QueryRow(ctx, "SHOW server_version").Scan(&pv)
	// if err != nil {
	// 	return nil, lazyerrors.Error(err)
	// }

	// pv, _, _ = strings.Cut(pv, " ")
	// mv := version.Get()

	// }

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{Documents: []*types.Document{resDoc}})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
