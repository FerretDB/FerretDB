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

package hanadb

import (
	"context"
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// CreateSchema creates a schema in SAP HANA JSON Document Store.
//
// Returns ErrSchemaAlreadyExist if schema already exists.
func (hanaPool *Pool) CreateSchema(ctx context.Context, qp *QueryParams) error {
	sqlStmt := fmt.Sprintf("CREATE SCHEMA %q", qp.DB)

	_, err := hanaPool.ExecContext(ctx, sqlStmt)

	return getHanaErrorIfExists(err)
}

// CreateSchemaIfNotExists creates a schema in SAP HANA JSON Document Store.
//
// Returns nil if the schema already exists.
func (hanaPool *Pool) CreateSchemaIfNotExists(ctx context.Context, qp *QueryParams) error {
	err := hanaPool.CreateSchema(ctx, qp)

	switch {
	case errors.Is(err, ErrSchemaAlreadyExist):
		return nil
	default:
		return err
	}
}

// DropSchema drops database.
//
// Returns ErrSchemaNotExist if schema does not exist.
func (hanaPool *Pool) DropSchema(ctx context.Context, qp *QueryParams) error {
	sql := fmt.Sprintf("DROP SCHEMA %q CASCADE", qp.DB)

	_, err := hanaPool.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

// ListSchemas lists all schemas that aren't related to Hana SYS schemas and SYS owner.
func (hanaPool *Pool) ListSchemas(ctx context.Context) ([]string, error) {
	const excludeSYS = "%SYS%"
	sqlStmt := "SELECT SCHEMA_NAME FROM SCHEMAS WHERE SCHEMA_NAME NOT LIKE $1 AND SCHEMA_OWNER NOT LIKE $2"

	rows, err := hanaPool.QueryContext(ctx, sqlStmt, excludeSYS, excludeSYS)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	res := make([]string, 0, 2)

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, lazyerrors.Error(err)
		}

		res = append(res, name)
	}

	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// ListCollections lists all collections under a given schema.
//
// Returns an empty array if schema doesn't exist.
func (hanaPool *Pool) ListCollections(ctx context.Context, qp *QueryParams) ([]string, error) {
	sqlStmt := "SELECT TABLE_NAME FROM M_TABLES WHERE SCHEMA_NAME = $1 AND TABLE_TYPE = 'COLLECTION'"

	rows, err := hanaPool.QueryContext(ctx, sqlStmt, qp.DB)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	res := make([]string, 0, 2)

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, lazyerrors.Error(err)
		}

		res = append(res, name)
	}

	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// SchemaSize calculates the total size of collections under schema.
//
// Returns 0 if schema doesn't exist, otherwise returns its size.
func (hanaPool *Pool) SchemaSize(ctx context.Context, qp *QueryParams) (int64, error) {
	collections, err := hanaPool.ListCollections(ctx, qp)
	if err != nil {
		return 0, lazyerrors.Error(err)
	}

	qpCopy := QueryParams{
		DB: qp.DB,
	}

	var totalSize int64

	for _, collection := range collections {
		qpCopy.Collection = collection
		size, err := hanaPool.CollectionSize(ctx, &qpCopy)

		if err != nil {
			return 0, lazyerrors.Error(err)
		}

		totalSize += size
	}

	return totalSize, nil
}
