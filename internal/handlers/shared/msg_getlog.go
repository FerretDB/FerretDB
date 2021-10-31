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

func (h *Handler) MsgGetLog(ctx context.Context, header *wire.MsgHeader, msg *wire.OpMsg) (*wire.OpMsg, error) {
	if len(msg.Documents) != 1 {
		return nil, common.NewError(common.ErrNotImplemented, fmt.Errorf("multiple documents are not supported"), header, msg)
	}
	document := msg.Documents[0]

	if document.Map()["getLog"] != "startupWarnings" {
		return nil, common.NewError(common.ErrNotImplemented, nil, header, msg)
	}

	var pv string
	err := h.pgPool.QueryRow(ctx, "SHOW server_version").Scan(&pv)
	if err != nil {
		return nil, err
	}

	pv = strings.Split(pv, " ")[0]
	mv := version.Get()

	b, err := json.Marshal(map[string]interface{}{
		"msg":  "Powered by 🥭 MangoDB " + mv.Version + " and PostgreSQL " + pv + ".",
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

	reply := &wire.OpMsg{
		Documents: []types.Document{types.MakeDocument(
			"totalLinesWritten", int32(1),
			"log", types.Array{string(b)},
			"ok", float64(1),
		)},
	}
	return reply, nil
}
