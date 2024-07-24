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
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/SAP/go-hdb/driver"
	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// collection implements backends.Collection interface by delegating all methods to the wrapped database.
// A collection in HANA is either stored in a table or as a column of a table. The column representation
// not supported yet.
type collection struct {
	hdb      *fsql.DB
	database string
	name     string
}

// newCollection creates a new Collection.
func newCollection(hdb *fsql.DB, database, name string) backends.Collection {
	return backends.CollectionContract(&collection{
		hdb:      hdb,
		database: database,
		name:     name,
	})
}

// generateQuery generates the HANA SQL query for the given filter and sort.
func (c *collection) generateQuery(filter, sort *types.Document) (string, error) {
	selectClause := prepareSelectClause(c.database, c.name)

	whereClause, err := prepareWhereClause(c.name, filter)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	sql := selectClause + whereClause

	orderByClause, err := prepareOrderByClause(sort)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	sql += orderByClause

	return sql, nil
}

// Query implements backends.Collection interface.
func (c *collection) Query(ctx context.Context, params *backends.QueryParams) (*backends.QueryResult, error) {
	db, err := databaseExists(ctx, c.hdb, c.database)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	if !db {
		return &backends.QueryResult{
			Iter: newQueryIterator(ctx, nil),
		}, nil
	}

	col, err := collectionExists(ctx, c.hdb, c.database, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !col {
		return &backends.QueryResult{
			Iter: newQueryIterator(ctx, nil),
		}, nil
	}

	sql, err := c.generateQuery(params.Filter, params.Sort)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

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
	err := createDatabaseIfNotExists(ctx, c.hdb, c.database)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	err = createCollectionIfNotExists(ctx, c.hdb, c.database, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	insertSQL := "INSERT INTO %q.%q values('%s')"

	for _, doc := range params.Docs {
		jsonBytes, err := marshalHana(doc)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		line := fmt.Sprintf(insertSQL, c.database, c.name, jsonBytes)

		_, err = c.hdb.ExecContext(ctx, line)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return new(backends.InsertAllResult), nil
}

// UpdateAll implements backends.Collection interface.
func (c *collection) UpdateAll(ctx context.Context, params *backends.UpdateAllParams) (*backends.UpdateAllResult, error) {
	db, err := databaseExists(ctx, c.hdb, c.database)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res backends.UpdateAllResult
	if !db {
		return &res, nil
	}

	col, err := collectionExists(ctx, c.hdb, c.database, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !col {
		return &res, nil
	}

	updateSQL := "UPDATE %q.%q SET %q = parse_json('%s') WHERE \"_id\" = %s"

	for _, doc := range params.Docs {
		jsonBytes, err := marshalHana(doc)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		id, _ := doc.Get("_id")
		must.NotBeZero(id)

		id = jsonToHanaQueryString(string(must.NotFail(sjson.MarshalSingleValue(id))))

		line := fmt.Sprintf(updateSQL, c.database, c.name, c.name, jsonBytes, id)

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
	db, err := databaseExists(ctx, c.hdb, c.database)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !db {
		return &backends.DeleteAllResult{Deleted: 0}, nil
	}

	col, err := collectionExists(ctx, c.hdb, c.database, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !col {
		return &backends.DeleteAllResult{Deleted: 0}, nil
	}

	var res backends.DeleteAllResult
	deleteSQL := "DELETE FROM %q.%q WHERE \"_id\" = %s"

	for _, id := range params.IDs {
		idString := jsonToHanaQueryString(string(must.NotFail(sjson.MarshalSingleValue(id))))
		line := fmt.Sprintf(deleteSQL, c.database, c.name, idString)

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
	// HANATODO HANA does not provide explain plan in json format. We need a conversion here.

	db, err := databaseExists(ctx, c.hdb, c.database)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !db {
		return &backends.ExplainResult{
			QueryPlanner: must.NotFail(types.NewDocument()),
		}, nil
	}

	col, err := collectionExists(ctx, c.hdb, c.database, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !col {
		return &backends.ExplainResult{
			QueryPlanner: must.NotFail(types.NewDocument()),
		}, nil
	}

	// Generate uuid for explain plan statement name.
	explainUUID := c.database + "_" + c.name + "_" + uuid.New().String()

	querySQL, err := c.generateQuery(params.Filter, params.Sort)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	explainSQL := fmt.Sprintf("EXPLAIN PLAN SET STATEMENT_NAME = '%s' FOR %s", explainUUID, querySQL)

	_, err = c.hdb.ExecContext(ctx, explainSQL)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	selectExplainSQL := fmt.Sprintf(
		"SELECT OPERATOR_NAME, OPERATOR_DETAILS, OPERATOR_ID, COALESCE(PARENT_OPERATOR_ID,0) "+
			"FROM EXPLAIN_PLAN_TABLE WHERE STATEMENT_NAME = '%s'", explainUUID,
	)

	rows, err := c.hdb.QueryContext(ctx, selectExplainSQL)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer rows.Close()

	// Build a document from the explain plan.
	var explainDoc types.Document

	for rows.Next() {
		var operatorName string
		var lobDetail driver.Lob
		storageDetails := new(bytes.Buffer)

		lobDetail.SetWriter(storageDetails)

		var operatorID, operatorParentID int32

		err = rows.Scan(&operatorName, &lobDetail, &operatorID, &operatorParentID)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		operatorDetails := storageDetails.String()

		explainDoc.Set(operatorName, must.NotFail(types.NewDocument(
			"operator_id", operatorID,
			"parent_operator_id", operatorParentID,
			"operator_details", operatorDetails,
		)))
	}

	cleanupSQL := fmt.Sprintf("DELETE FROM EXPLAIN_PLAN_TABLE WHERE STATEMENT_NAME = '%s'", explainUUID)

	_, err = c.hdb.ExecContext(ctx, cleanupSQL)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &backends.ExplainResult{
		QueryPlanner: &explainDoc,
	}, nil
}

// Stats implements backends.Collection interface.
func (c *collection) Stats(ctx context.Context, params *backends.CollectionStatsParams) (*backends.CollectionStatsResult, error) {
	var res backends.CollectionStatsResult

	db, err := databaseExists(ctx, c.hdb, c.database)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !db {
		return nil, backends.NewError(
			backends.ErrorCodeDatabaseDoesNotExist,
			lazyerrors.Errorf("no ns %s.%s", c.database, c.name),
		)
	}

	col, err := collectionExists(ctx, c.hdb, c.database, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !col {
		return nil, backends.NewError(
			backends.ErrorCodeCollectionDoesNotExist,
			lazyerrors.Errorf("no ns %s.%s", c.database, c.name),
		)
	}

	// HANATODO Fill out collection stats

	return &res, nil
}

// Compact implements backends.Collection interface.
func (c *collection) Compact(ctx context.Context, params *backends.CompactParams) (*backends.CompactResult, error) {
	// HANA does not provide compact functionality.
	return new(backends.CompactResult), nil
}

// Prefixes an index with the collection name to store it in hana.
//
// Reasoning:
// Hana DocStore stores indexes on a schema/database level, that may cause
// confilts for indexes of different collections having the same name (eg col1.idx and col2.idx).
func (c *collection) prefixIndexName(indexName string) string {
	return fmt.Sprintf("%s__%s", c.name, indexName)
}

// Removes the prefix of an index name.
func (c *collection) removeIndexNamePrefix(prefixedIndex string) string {
	prefix := fmt.Sprintf("%s__", c.name)
	return strings.Replace(prefixedIndex, prefix, "", 1)
}

// ListIndexes implements backends.Collection interface.
func (c *collection) ListIndexes(ctx context.Context, params *backends.ListIndexesParams) (*backends.ListIndexesResult, error) {
	db, err := databaseExists(ctx, c.hdb, c.database)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !db {
		return nil, backends.NewError(
			backends.ErrorCodeDatabaseDoesNotExist,
			lazyerrors.Errorf("no ns %s.%s", c.database, c.name),
		)
	}

	col, err := collectionExists(ctx, c.hdb, c.database, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !col {
		return nil, backends.NewError(
			backends.ErrorCodeCollectionDoesNotExist,
			lazyerrors.Errorf("no ns %s.%s", c.database, c.name),
		)
	}

	sql := "SELECT idx.INDEX_NAME, idx.COLUMN_NAME, idx.ASCENDING_ORDER " +
		"FROM INDEX_COLUMNS idx, M_TABLES tbl " +
		"WHERE idx.SCHEMA_NAME = '%s' AND idx.TABLE_NAME = '%s' AND " +
		"idx.SCHEMA_NAME = tbl.SCHEMA_NAME AND idx.TABLE_NAME = tbl.TABLE_NAME AND " +
		"tbl.TABLE_TYPE = 'COLLECTION'"

	sql = fmt.Sprintf(sql, c.database, c.name)

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
			Name:   c.removeIndexNamePrefix(idxName),
			Unique: true, // HANATODO: Is this possible to query from HANA?
			Key:    []backends.IndexKeyPair{{Field: idxColName, Descending: !ascending}},
		})
	}

	res := backends.ListIndexesResult{Indexes: indexes}

	sort.Slice(res.Indexes, func(i, j int) bool {
		return res.Indexes[i].Name < res.Indexes[j].Name
	})

	return &res, nil
}

// indexExists checks if an index exists in a list of indexes retrieved from HANA.
func indexExists(indexes []backends.IndexInfo, indexToFind string) bool {
	for _, idx := range indexes {
		if idx.Name == indexToFind {
			return true
		}
	}

	return false
}

// CreateIndexes implements backends.Collection interface.
func (c *collection) CreateIndexes(ctx context.Context, params *backends.CreateIndexesParams) (*backends.CreateIndexesResult, error) { //nolint:lll // for readability
	db, err := databaseExists(ctx, c.hdb, c.database)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !db {
		panic("database does not exist")
	}

	err = createCollectionIfNotExists(ctx, c.hdb, c.database, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	existingIndexes, err := c.ListIndexes(ctx, new(backends.ListIndexesParams))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sql := "CREATE HASH INDEX %q.%q ON %q.%q(%q)"

	var createStmt string

	// HANATODO Can we support more than one field for indexes in HANA DocStore?
	for _, index := range params.Indexes {
		if !indexExists(existingIndexes.Indexes, index.Name) {
			createStmt = fmt.Sprintf(sql, c.database, c.prefixIndexName(index.Name), c.database, c.name, index.Key[0].Field)

			_, err := c.hdb.ExecContext(ctx, createStmt)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
	}

	return new(backends.CreateIndexesResult), nil
}

// DropIndexes implements backends.Collection interface.
func (c *collection) DropIndexes(ctx context.Context, params *backends.DropIndexesParams) (*backends.DropIndexesResult, error) {
	_, err := databaseExists(ctx, c.hdb, c.database)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// HANATODO Check if index is on this collection.

	sql := "DROP INDEX %q.%q"

	var droptStmt string

	for _, index := range params.Indexes {
		droptStmt = fmt.Sprintf(sql, c.database, c.prefixIndexName(index))

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
