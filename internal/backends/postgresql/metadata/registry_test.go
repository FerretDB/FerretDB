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

package metadata

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// testCollection creates, tests, and drops an unique collection in existing database.
func testCollection(t *testing.T, ctx context.Context, r *Registry, db *pgxpool.Pool, dbName, collectionName string) {
	t.Helper()

	c := r.CollectionGet(ctx, dbName, collectionName)
	require.Nil(t, c)

	created, err := r.CollectionCreate(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.True(t, created)

	created, err = r.CollectionCreate(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.False(t, created)

	c = r.CollectionGet(ctx, dbName, collectionName)
	require.NotNil(t, c)
	require.Equal(t, collectionName, c.Name)

	list, err := r.CollectionList(ctx, dbName)
	require.NoError(t, err)
	require.Contains(t, list, c)

	q := fmt.Sprintf(`INSERT INTO %s (%s) VALUES($1)`, pgx.Identifier{dbName, c.TableName}.Sanitize(), DefaultColumn)
	doc := `{"$s": {"p": {"_id": {"t": "int"}}, "$k": ["_id"]}, "_id": 42}`
	_, err = db.Exec(ctx, q, doc)
	require.NoError(t, err)

	dropped, err := r.CollectionDrop(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.True(t, dropped)

	dropped, err = r.CollectionDrop(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.False(t, dropped)

	c = r.CollectionGet(ctx, dbName, collectionName)
	require.Nil(t, c)
}

func TestCreateDrop(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)

	connInfo := conninfo.NewConnInfo()
	t.Cleanup(connInfo.Close)

	ctx = conninfo.WithConnInfo(ctx, connInfo)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	r, err := NewRegistry("postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1", testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName := testutil.DatabaseName(t)

	db, err := r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)

	collectionName := testutil.CollectionName(t)

	testCollection(t, ctx, r, db, dbName, collectionName)
}
