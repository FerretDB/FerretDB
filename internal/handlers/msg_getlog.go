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

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/version"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetLog is an administrative command that returns the most recent 1024 logged events.
func (h *Handler) MsgGetLog(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if l := document.Map()["getLog"]; l != "startupWarnings" {
		errMsg := fmt.Sprintf("MsgGetLog: unhandled getLog value %q", l)
		return nil, common.NewErrorMsg(common.ErrNotImplemented, errMsg)
	}

	var pv string
	err = h.pgPool.QueryRow(ctx, "SHOW server_version").Scan(&pv)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	pv, _, _ = strings.Cut(pv, " ")
	mv := version.Get()

	var log types.Array
	for _, line := range []string{
		"Powered by 🥭 FerretDB " + mv.Version + " and PostgreSQL " + pv + ".",
		"Please star us on GitHub: https://github.com/FerretDB/FerretDB",
	} {
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
		if err = log.Append(string(b)); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"totalLinesWritten", int32(log.Len()),
			"log", &log,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
