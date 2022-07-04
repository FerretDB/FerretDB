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

package pgdb

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestQueryDocuments(t *testing.T) {
	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t, nil, zaptest.NewLogger(t))

	schemaName := testutil.SchemaName(t)
	tableName := testutil.TableName(t)

	err := pool.CreateDatabase(ctx, schemaName)
	require.NoError(t, err)

	err = pool.CreateCollection(ctx, schemaName, tableName)
	require.NoError(t, err)

	// 0 docs
	// 1 doc
	// 2 docs
	// 3 docs

	// chan full

	//	pool.InsertDocument(ctx, schemaName, tableName,)

	//	pool.QueryDocuments()
}
