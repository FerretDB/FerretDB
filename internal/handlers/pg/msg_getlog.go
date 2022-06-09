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

	"go.uber.org/zap/zapcore"

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

	// "Powered by 🥭 FerretDB " + mv.Version + " and PostgreSQL " + pv + ".",
	// "Please star us on GitHub: https://github.com/FerretDB/FerretDB",

	command := document.Command()

	getLog, err := document.Get(command)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, ok := getLog.(string); !ok {
		return nil, common.NewErrorMsg(common.ErrTypeMismatch, "Argument to getLog must be of type String")
	}

	var resDoc *types.Document
	switch getLog {
	case "*":
		resDoc = must.NotFail(types.NewDocument(
			"names", must.NotFail(types.NewArray("global", "startupWarnings")),
			"ok", float64(1),
		))

	case "global":
		log, err := requirRecordsLog(zapcore.DebugLevel)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		resDoc = must.NotFail(types.NewDocument(
			"log", &log,
			"totalLinesWritten", int64(log.Len()),
			"ok", float64(1),
		))

	case "startupWarnings":
		log, err := requirRecordsLog(zapcore.WarnLevel)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		resDoc = must.NotFail(types.NewDocument(
			"log", &log,
			"totalLinesWritten", int64(log.Len()),
			"ok", float64(1),
		))

	default:
		errMsg := fmt.Sprintf("no RamLog named: %s", getLog)
		return nil, common.NewErrorMsg(0, errMsg)
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{Documents: []*types.Document{resDoc}})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// requirRecordsLog returns an array of records from logging buffer with given level.
func requirRecordsLog(level zapcore.Level) (log types.Array, err error) {
	entries := logging.RecentEntries.Get()
	for _, e := range entries {
		if e.Level >= level {
			b, err := json.Marshal(map[string]any{
				"t":   e.Time,
				"l":   e.Level,
				"ln":  e.LoggerName,
				"msg": e.Message,
				"c":   e.Caller,
				"s":   e.Stack,
			})
			if err != nil {
				return types.Array{}, err
			}
			if err = log.Append(string(b)); err != nil {
				return types.Array{}, err
			}
		}
	}

	return log, nil
}
