// Copyright 2021 Baltoro OÜ.
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

package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MangoDB-io/MangoDB/internal/handlers/common"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/version"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

func (h *Handler) MsgGetLog(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, common.NewError(common.ErrInternalError, err)
	}

	if l := document.Map()["getLog"]; l != "startupWarnings" {
		return nil, common.NewError(common.ErrNotImplemented, fmt.Errorf("unhandled getLog value %q", l))
	}

	var pv string
	err = h.pgPool.QueryRow(ctx, "SHOW server_version").Scan(&pv)
	if err != nil {
		return nil, err
	}

	pv = strings.Split(pv, " ")[0]
	mv := version.Get()

	var log types.Array
	for _, line := range []string{
		"Powered by 🥭 MangoDB " + mv.Version + " and PostgreSQL " + pv + ".",
		"Please star us on GitHub: https://github.com/MangoDB-io/MangoDB",
	} {
		b, err := json.Marshal(map[string]interface{}{
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
			return nil, err
		}
		log = append(log, string(b))
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"totalLinesWritten", int32(len(log)),
			"log", log,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, common.NewError(common.ErrInternalError, err)
	}

	return &reply, nil
}
