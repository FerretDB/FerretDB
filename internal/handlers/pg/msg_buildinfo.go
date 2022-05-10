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
	"strconv"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/version"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// For clients that check version.
const versionValue = "5.0.42"

// MsgBuildInfo returns an OpMsg with the build information.
func (h *Handler) MsgBuildInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	var reply wire.OpMsg
	err := reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"version", versionValue,
			"gitVersion", version.Get().Commit,
			"modules", must.NotFail(types.NewArray()),
			"sysInfo", "deprecated",
			"versionArray", must.NotFail(types.NewArray(int32(5), int32(0), int32(42), int32(0))),
			"bits", int32(strconv.IntSize),
			"debug", version.Get().Debug,
			"maxBsonObjectSize", int32(types.MaxDocumentLen),
			"buildEnvironment", version.Get().BuildEnvironment,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
