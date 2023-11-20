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

package hana

import (
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// collection implements backends.Collection interface by delegating all methods to the wrapped database.
// A collection in HANA is either stored in a table or as a column of a table. The column representation
// not supported yet.
type collection struct {
	hdb    *fsql.DB
	schema string
	table  string
	// column string
}

// newCollection creates a new Collection.
func newCollection(hdb *fsql.DB, schema, table string) backends.Collection {
	return backends.CollectionContract(&collection{
		hdb:    hdb,
		schema: schema,
		table:  table,
	})
}

// Query implements backends.Collection interface.
func (c *collection) Query(ctx context.Context, params *backends.QueryParams) (*backends.QueryResult, error) {

	var args []any

	selectClause, selectArgs := prepareSelectClause(c.schema, c.table)
	args = append(args, selectArgs...)

	whereClause, whereArgs, err := prepareWhereClause(params.Filter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sql := selectClause + whereClause
	args = append(args, whereArgs...)

	orderByClause, orderByArgs, err := prepareOrderByClause(params.Sort)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sql += orderByClause
	args = append(args, orderByArgs...)

	sql = fmt.Sprintf(sql, args...)

	rows, err := c.hdb.QueryContext(ctx, sql)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &backends.QueryResult{
		Iter: newQueryIterator(ctx, rows),
	}, nil
}

// InsertAll implements backends.Collection interface.
func (c *collection) InsertAll(ctx context.Context, params *backends.InsertAllParams) (*backends.InsertAllResult, error) {
	// TODO: Create schema&collection if not exists.

	insertSql := "INSERT INTO %q.%q values('%s')"

	for _, doc := range params.Docs {
		jsonBytes, err := MarshalHana(doc)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		line := fmt.Sprintf(insertSql, c.schema, c.table, jsonBytes)
		_, err = c.hdb.ExecContext(ctx, line)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return new(backends.InsertAllResult), nil
}

// UpdateAll implements backends.Collection interface.
func (c *collection) UpdateAll(ctx context.Context, params *backends.UpdateAllParams) (*backends.UpdateAllResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// DeleteAll implements backends.Collection interface.
func (c *collection) DeleteAll(ctx context.Context, params *backends.DeleteAllParams) (*backends.DeleteAllResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// Explain implements backends.Collection interface.
func (c *collection) Explain(ctx context.Context, params *backends.ExplainParams) (*backends.ExplainResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// Stats implements backends.Collection interface.
func (c *collection) Stats(ctx context.Context, params *backends.CollectionStatsParams) (*backends.CollectionStatsResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// Compact implements backends.Collection interface.
func (c *collection) Compact(ctx context.Context, params *backends.CompactParams) (*backends.CompactResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// ListIndexes implements backends.Collection interface.
func (c *collection) ListIndexes(ctx context.Context, params *backends.ListIndexesParams) (*backends.ListIndexesResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// CreateIndexes implements backends.Collection interface.
func (c *collection) CreateIndexes(ctx context.Context, params *backends.CreateIndexesParams) (*backends.CreateIndexesResult, error) { //nolint:lll // for readability
	return nil, lazyerrors.New("not implemented yet")
}

// DropIndexes implements backends.Collection interface.
func (c *collection) DropIndexes(ctx context.Context, params *backends.DropIndexesParams) (*backends.DropIndexesResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)
