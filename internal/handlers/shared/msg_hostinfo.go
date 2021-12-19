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

package shared

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// Now is a synonym for the time.Now function (very handy for unit testing)
var Now = time.Now

// MsgHostInfo returns an OpMsg with the host information.
func (h *Handler) MsgHostInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	now := Now().UTC().Format(time.RFC3339)
	hostname, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)

	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"system", types.MustMakeDocument(
				"currentTime", fmt.Sprintf("ISODate(%s)", now),
				"hostname", hostname,
				"cpuAddrSize", fmt.Sprintf("%d", strconv.IntSize),
				"numCores", fmt.Sprintf("%d", runtime.NumCPU()),
				"cpuArch", runtime.GOARCH,
				"numaEnabled", "false",
			),
			"os", types.MustMakeDocument(
				"type", strings.ToTitle(runtime.GOOS),
			),
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
