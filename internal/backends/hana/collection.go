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

	insertSQL := "INSERT INTO %q.%q VALUES('%s')"

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

	updateSQL := "UPDATE %q.%q SET %q = parse_json('%s') WHERE \"_id\" = '%s'"

	for _, doc := range params.Docs {
		jsonBytes, err := marshalHana(doc)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		id, _ := doc.Get("_id")
		must.NotBeZero(id)

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
			backends.ErrorCodeCollectionDoesNotExist,
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

	queryCountDocuments := "SELECT count(*) FROM %q.%q"
	queryCountDocuments = fmt.Sprintf(queryCountDocuments, c.database, c.name)

	countDocuments, err := querySingleInt(queryCountDocuments, ctx, c.hdb)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	querySizeTotal := "SELECT COALESCE(SUM(TABLE_SIZE),0) FROM M_TABLES " +
		"WHERE SCHEMA_NAME = '%s' AND TABLE_NAME = '%s'"
	querySizeTotal = fmt.Sprintf(querySizeTotal, c.database, c.name)

	sizeTotal, err := querySingleInt(querySizeTotal, ctx, c.hdb)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	queryFreeMemory := "SELECT FREE_PHYSICAL_MEMORY  FROM M_HOST_RESOURCE_UTILIZATION"

	freeMemory, err := querySingleInt(queryFreeMemory, ctx, c.hdb)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res.CountDocuments = countDocuments
	res.SizeCollection = sizeTotal
	res.SizeIndexes = 1 // see database.Stats
	res.SizeTotal = res.SizeCollection + res.SizeIndexes
	// Note: this does currently not take capped collections into account.
	res.SizeFreeStorage = freeMemory

	return &res, nil
}

// Compact implements backends.Collection interface.
func (c *collection) Compact(ctx context.Context, params *backends.CompactParams) (*backends.CompactResult, error) {
	// HANA does not provide compact functionality.
	return new(backends.CompactResult), nil
}

// ListIndexes implements backends.Collection interface.
func (c *collection) ListIndexes(ctx context.Context, params *backends.ListIndexesParams) (*backends.ListIndexesResult, error) {
	return listIndexes(ctx, c.hdb, c.database, c.name)
}

// CreateIndexes implements backends.Collection interface.
func (c *collection) CreateIndexes(ctx context.Context, params *backends.CreateIndexesParams) (*backends.CreateIndexesResult, error) { //nolint:lll // for readability
	return createIndexes(ctx, c.hdb, c.database, c.name, params)
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
		droptStmt = fmt.Sprintf(sql, c.database, prefixIndexName(c.name, index))

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
