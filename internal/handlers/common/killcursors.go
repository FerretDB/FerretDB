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

package common

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// KillCursors is a part of common implementation of the killCursors command.
func KillCursors(ctx context.Context, msg *wire.OpMsg, registry *cursor.Registry) (*wire.OpMsg, error) {
	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursorsKilled", types.MakeArray(0),
			"cursorsNotFound", types.MakeArray(0),
			"cursorsAlive", types.MakeArray(0),
			"cursorsUnknown", types.MakeArray(0),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
