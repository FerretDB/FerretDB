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
	"sort"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/sjson"
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

// Helperfunction to check if database and collection exist on hana.
func (c *collection) checkSchemaAndCollectionExists(ctx context.Context) error {
	s, err := SchemaExists(ctx, c.hdb, c.schema)
	if !s {
		return lazyerrors.Errorf("Database %q does not exist!", c.schema)
	}
	if err != nil {
		return lazyerrors.Error(err)
	}

	col, err := CollectionExists(ctx, c.hdb, c.schema, c.table)
	if !col {
		return lazyerrors.Errorf("Collection %q does not exist!", c.table)
	}
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// Query implements backends.Collection interface.
func (c *collection) Query(ctx context.Context, params *backends.QueryParams) (*backends.QueryResult, error) {
	err := c.checkSchemaAndCollectionExists(ctx)
	if err != nil {
		return nil, err
	}

	selectClause := prepareSelectClause(c.schema, c.table)

	whereClause, err := prepareWhereClause(c.table, params.Filter)
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
	err := c.checkSchemaAndCollectionExists(ctx)
	if err != nil {
		return nil, err
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

	err := c.checkSchemaAndCollectionExists(ctx)
	if err != nil {
		return nil, err
	}

	var res backends.DeleteAllResult
	deleteSql := "DELETE FROM %q.%q WHERE \"_id\" = %s"

	for _, id := range params.IDs {
		idString := jsonToHanaQueryString(string(must.NotFail(sjson.MarshalSingleValue(id))))
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
	// TODO HANA does not provide explain plan in json format. We need a conversion here.
	return nil, lazyerrors.New("not implemented yet")
}

// Stats implements backends.Collection interface.
func (c *collection) Stats(ctx context.Context, params *backends.CollectionStatsParams) (*backends.CollectionStatsResult, error) {
	var res backends.CollectionStatsResult
	err := c.checkSchemaAndCollectionExists(ctx)
	if err != nil {
		return nil, err
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
	err := c.checkSchemaAndCollectionExists(ctx)
	if err != nil {
		return nil, err
	}
	sql := "SELECT idx.INDEX_NAME, idx.COLUMN_NAME, idx.ASCENDING_ORDER " +
		"FROM INDEX_COLUMNS idx, M_TABLES tbl " +
		"WHERE idx.SCHEMA_NAME = '%s' AND idx.TABLE_NAME = '%s' AND " +
		"idx.SCHEMA_NAME = tbl.SCHEMA_NAME AND idx.TABLE_NAME = tbl.TABLE_NAME AND " +
		"tbl.TABLE_TYPE = 'COLLECTION'"

	sql = fmt.Sprintf(sql, c.schema, c.table)
	rows, err := c.hdb.QueryContext(ctx, sql)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer rows.Close()

	var indexes []backends.IndexInfo

	for rows.Next() {
		var idxName, idxColName string
		var ascending bool

		err = rows.Scan(&idxName, &idxColName, &ascending)
		if err != nil {
			return nil, lazyerrors.Error(nil)
		}

		indexes = append(indexes, backends.IndexInfo{
			Name:   idxName,
			Unique: true, // TODO: Is this possible to query from HANA?
			Key:    []backends.IndexKeyPair{{Field: idxColName, Descending: !ascending}},
		})
	}

	res := backends.ListIndexesResult{Indexes: indexes}

	sort.Slice(res.Indexes, func(i, j int) bool {
		return res.Indexes[i].Name < res.Indexes[j].Name
	})

	return &res, nil
}

// CreateIndexes implements backends.Collection interface.
func (c *collection) CreateIndexes(ctx context.Context, params *backends.CreateIndexesParams) (*backends.CreateIndexesResult, error) { //nolint:lll // for readability
	err := c.checkSchemaAndCollectionExists(ctx)
	if err != nil {
		return nil, err
	}

	sql := "CREATE HASH INDEX %q.%s ON %q.%q(%q)"
	var createStmt string
	// TODO Can we support more than one field for indexes in HANA?
	for _, index := range params.Indexes {
		createStmt = fmt.Sprintf(sql, c.schema, index.Name, c.schema, c.table, index.Key[0].Field)
		_, err := c.hdb.ExecContext(ctx, createStmt)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return new(backends.CreateIndexesResult), nil
}

// DropIndexes implements backends.Collection interface.
func (c *collection) DropIndexes(ctx context.Context, params *backends.DropIndexesParams) (*backends.DropIndexesResult, error) {
	err := c.checkSchemaAndCollectionExists(ctx)
	if err != nil {
		return nil, err
	}

	// TODO Check if index is on this collection.

	sql := "DROP INDEX %q.%s"
	var droptStmt string
	for _, index := range params.Indexes {
		droptStmt = fmt.Sprintf(sql, c.schema, index)
		_, err := c.hdb.ExecContext(ctx, droptStmt)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return new(backends.DropIndexesResult), nil
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)
