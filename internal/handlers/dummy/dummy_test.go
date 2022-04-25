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

package dummy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

func TestDummyHandler(t *testing.T) {
	t.Parallel()

	h := New()
	ctx := context.Background()
	var msg wire.OpMsg
	err := msg.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"commands", must.NotFail(types.NewDocument()),
			"ok", float64(1),
		))},
	})
	assert.NoError(t, err)

	errNotImplemented := common.NewErrorMsg(common.ErrNotImplemented, "I'm a dummy, not a handler")

	for _, fn := range []func(context.Context, *wire.OpMsg) (*wire.OpMsg, error){
		h.MsgBuildInfo,
		h.MsgCollStats,
		h.MsgCreate,
		h.MsgDataSize,
		h.MsgDBStats,
		h.MsgDrop,
		h.MsgDropDatabase,
		h.MsgGetCmdLineOpts,
		h.MsgGetLog,
		h.MsgGetParameter,
		h.MsgHostInfo,
		h.MsgIsMaster,
		h.MsgHello,
		h.MsgListCollections,
		h.MsgListDatabases,
		h.MsgPing,
		h.MsgServerStatus,
		h.MsgWhatsMyURI,
		h.MsgCount,
		h.MsgCreateIndexes,
		h.MsgDelete,
		h.MsgFind,
		h.MsgFindAndModify,
		h.MsgInsert,
		h.MsgUpdate,
		h.MsgDebugError,
		h.MsgDebugPanic,
	} {
		_, err := fn(ctx, &msg)
		assert.Equal(t, err, errNotImplemented)
	}

	msgq := new(wire.OpQuery)
	_, err = h.MsgQueryCmd(ctx, msgq)
	assert.Equal(t, err, errNotImplemented)
}
