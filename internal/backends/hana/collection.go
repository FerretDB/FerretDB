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
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
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
	s, err := SchemaExists(ctx, c.hdb, c.schema)
	if !s {
		return nil, lazyerrors.Errorf("Schema %q does not exist!", c.schema)
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	selectClause := prepareSelectClause(c.schema, c.table)

	whereClause, err := prepareWhereClause(params.Filter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sql := selectClause + whereClause

	orderByClause, err := prepareOrderByClause(params.Sort)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sql += orderByClause

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
	err := CreateSchemaIfNotExists(ctx, c.hdb, c.schema)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	err = CreateCollectionIfNotExists(ctx, c.hdb, c.schema, c.table)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

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
	var res backends.UpdateAllResult
	s, err := SchemaExists(ctx, c.hdb, c.schema)
	if !s {
		return &res, nil
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	col, err := CollectionExists(ctx, c.hdb, c.schema, c.table)
	if !col {
		return &res, nil
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	updateSql := "UPDATE %q.%q SET %q = (%s) WHERE \"_id\" = %q"

	for _, doc := range params.Docs {
		jsonBytes, err := MarshalHana(doc)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		id, _ := doc.Get("_id")
		must.NotBeZero(id)

		line := fmt.Sprintf(updateSql, c.schema, c.table, c.table, jsonBytes, id)
		execResult, err := c.hdb.ExecContext(ctx, line)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		numRows, err := execResult.RowsAffected()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res.Updated += int32(numRows)
	}

	return &res, nil
}

// DeleteAll implements backends.Collection interface.
func (c *collection) DeleteAll(ctx context.Context, params *backends.DeleteAllParams) (*backends.DeleteAllResult, error) {
	var res backends.DeleteAllResult
	s, err := SchemaExists(ctx, c.hdb, c.schema)
	if !s {
		return &res, nil
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	col, err := CollectionExists(ctx, c.hdb, c.schema, c.table)
	if !col {
		return &res, nil
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	deleteSql := "DELETE FROM %q.%q WHERE \"_id\" = '%s'"

	for _, id := range params.IDs {
		idString := string(must.NotFail(sjson.MarshalSingleValue(id)))
		line := fmt.Sprintf(deleteSql, c.schema, c.table, idString)
		execResult, err := c.hdb.ExecContext(ctx, line)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		numRows, err := execResult.RowsAffected()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res.Deleted += int32(numRows)
	}

	return &res, nil
}

// Explain implements backends.Collection interface.
func (c *collection) Explain(ctx context.Context, params *backends.ExplainParams) (*backends.ExplainResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// Stats implements backends.Collection interface.
func (c *collection) Stats(ctx context.Context, params *backends.CollectionStatsParams) (*backends.CollectionStatsResult, error) {
	var res backends.CollectionStatsResult
	s, err := SchemaExists(ctx, c.hdb, c.schema)
	if !s {
		return nil, backends.NewError(
			backends.ErrorCodeDatabaseDoesNotExist,
			lazyerrors.Errorf("No database (schema) with name %q", c.schema))
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	col, err := CollectionExists(ctx, c.hdb, c.schema, c.table)
	if !col {
		return nil, backends.NewError(
			backends.ErrorCodeCollectionDoesNotExist,
			lazyerrors.Errorf("No collection with name %q.%q", c.schema, c.table))
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO: Fill out collection stats

	return &res, nil
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
