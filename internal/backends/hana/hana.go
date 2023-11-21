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

// Package hana provides backend for SAP HANA.
package hana

import (
	"context"
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/SAP/go-hdb/driver"
) // register database/sql driver

// Errors are wrapped with lazyerrors.Error,
// so the caller needs to use errors.Is to check the error,
// for example, errors.Is(err, ErrSchemaNotExist).
var (
	// ErrTableNotExist indicates that there is no such table.
	ErrTableNotExist = fmt.Errorf("collection/table does not exist")

	// ErrSchemaNotExist indicates that there is no such schema.
	ErrSchemaNotExist = fmt.Errorf("database/schema does not exist")

	// ErrSchemaAlreadyExist indicates that a schema already exists.
	ErrSchemaAlreadyExist = fmt.Errorf("database/schema already exists")

	// ErrCollectionAlreadyExist indicates that a collection already exists.
	ErrCollectionAlreadyExist = fmt.Errorf("collection/table already exists")

	// ErrInvalidCollectionName indicates that a collection didn't pass name checks.
	ErrInvalidCollectionName = fmt.Errorf("invalid FerretDB collection name")

	// ErrInvalidDatabaseName indicates that a database didn't pass name checks.
	ErrInvalidDatabaseName = fmt.Errorf("invalid FerretDB database name")
)

// Errors from HanaDB.
var Errors = map[int]error{
	259: ErrTableNotExist,
	288: ErrCollectionAlreadyExist,
	362: ErrSchemaNotExist,
	386: ErrSchemaAlreadyExist,
}

// getHanaErrorIfExists converts 'err' to formatted version if it exists
//
// Returns one of the errors above or the original error if the formatted version doesn't exist.
func getHanaErrorIfExists(err error) error {
	var dbError driver.Error
	if errors.As(err, &dbError) {
		if hanaErr, ok := Errors[dbError.Code()]; ok {
			return hanaErr
		}
	}

	return err
}

func SchemaExists(ctx context.Context, hdb *fsql.DB, schema string) (bool, error) {
	sql := "SELECT COUNT(*) FROM \"PUBLIC\".\"SCHEMAS\" WHERE SCHEMA_NAME = ?"
	var count int

	if err := hdb.QueryRowContext(ctx, sql, schema).Scan(&count); err != nil {
		return false, lazyerrors.Error(err)
	}

	return count == 1, nil
}

// CreateSchema creates a schema in SAP HANA JSON Document Store.
//
// Returns ErrSchemaAlreadyExist if schema already exists.
func CreateSchema(ctx context.Context, hdb *fsql.DB, schema string) error {
	sqlStmt := fmt.Sprintf("CREATE SCHEMA %q", schema)

	_, err := hdb.ExecContext(ctx, sqlStmt)

	return getHanaErrorIfExists(err)
}

// CreateSchema creates a schema in SAP HANA JSON Document Store if it not exists.
func CreateSchemaIfNotExists(ctx context.Context, hdb *fsql.DB, schema string) error {
	err := CreateSchema(ctx, hdb, schema)

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
func DropSchema(ctx context.Context, hdb *fsql.DB, schema string) error {
	sql := fmt.Sprintf("DROP SCHEMA %q CASCADE", schema)

	_, err := hdb.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

func CollectionExists(ctx context.Context, hdb *fsql.DB, schema, table string) (bool, error) {
	sql := "SELECT count(*) FROM M_TABLES WHERE TABLE_TYPE = 'COLLECTION' AND SCHEMA_NAME = ? AND TABLE_NAME = ?"
	var count int

	if err := hdb.QueryRowContext(ctx, sql, schema, table).Scan(&count); err != nil {
		return false, lazyerrors.Error(err)
	}

	return count == 1, nil
}

// CreateCollection creates a new SAP HANA JSON Document Store collection.
//
// It returns ErrAlreadyExist if collection already exist.
func CreateCollection(ctx context.Context, hdb *fsql.DB, schema, table string) error {
	sql := fmt.Sprintf("CREATE COLLECTION %q.%q", schema, table)

	_, err := hdb.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

// CreateCollectionIfNotExists creates a new SAP HANA JSON Document Store collection.
//
// Returns nil if collection already exist.
func CreateCollectionIfNotExists(ctx context.Context, hdb *fsql.DB, schema, table string) error {
	err := CreateCollection(ctx, hdb, schema, table)

	switch {
	case errors.Is(err, ErrCollectionAlreadyExist):
		return nil
	default:
		return err
	}
}
