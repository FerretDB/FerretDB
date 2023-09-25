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

package postgresql

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	params := NewBackendParams{
		URI: "postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1",
		L:   testutil.Logger(t),
		P:   sp,
	}
	b, err := NewBackend(&params)
	require.NoError(t, err)

	defer b.Close()

	db, err := b.Database(testutil.DatabaseName(t))
	require.NoError(t, err)

	c, err := db.Collection(testutil.CollectionName(t))
	require.NoError(t, err)

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	doc, err := types.NewDocument("_id", types.NewObjectID())
	require.NoError(t, err)

	_, err = c.InsertAll(ctx, &backends.InsertAllParams{
		Docs: []*types.Document{doc},
	})
	require.NoError(t, err)
	// TODO https://github.com/FerretDB/FerretDB/issues/3399
	//_, err = c.InsertAll(ctx, &backends.InsertAllParams{
	//	Docs: []*types.Document{doc},
	//})
	//require.True(t, backends.ErrorCodeIs(err, backends.ErrorCodeInsertDuplicateID))
}
