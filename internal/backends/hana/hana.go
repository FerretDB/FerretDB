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

	// Register HANA SQL driver.
	"github.com/SAP/go-hdb/driver"

	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

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

// hanaErrorCollectionNotExist checks if the error is about collection not existing.
//
// Returns true if the err is collection does not exist.
func hanaErrorCollectionNotExist(err error) bool {
	var dbError driver.Error
	if errors.As(err, &dbError) {
		return dbError.Code() == 259
	}

	return false
}

// hanaErrorSchemaNotExist checks if the error is about collection not existing.
//
// Returns true if the err is schema does not exist.
func hanaErrorSchemaNotExist(err error) bool {
	var dbError driver.Error
	if errors.As(err, &dbError) {
		return dbError.Code() == 362
	}

	return false
}

func schemaExists(ctx context.Context, hdb *fsql.DB, schema string) (bool, error) {
	sql := fmt.Sprintf("SELECT COUNT(*) FROM \"PUBLIC\".\"SCHEMAS\" WHERE SCHEMA_NAME = '%s'", schema)

	var count int
	if err := hdb.QueryRowContext(ctx, sql).Scan(&count); err != nil {
		return false, lazyerrors.Error(err)
	}

	return count == 1, nil
}

// CreateSchema creates a schema in SAP HANA JSON Document Store.
//
// Returns ErrSchemaAlreadyExist if schema already exists.
func createSchema(ctx context.Context, hdb *fsql.DB, schema string) error {
	sqlStmt := fmt.Sprintf("CREATE SCHEMA %q", schema)

	_, err := hdb.ExecContext(ctx, sqlStmt)

	return getHanaErrorIfExists(err)
}

func createSchemaIfNotExists(ctx context.Context, hdb *fsql.DB, schema string) error {
	exists, err := schemaExists(ctx, hdb, schema)
	if err != nil {
		return getHanaErrorIfExists(err)
	}

	if !exists {
		err = createSchema(ctx, hdb, schema)
	}

	return err
}

// DropSchema drops database.
//
// Returns ErrSchemaNotExist if schema does not exist.
func dropSchema(ctx context.Context, hdb *fsql.DB, schema string) (bool, error) {
	sql := fmt.Sprintf("DROP SCHEMA %q CASCADE", schema)

	_, err := hdb.ExecContext(ctx, sql)
	if err != nil {
		if hanaErrorSchemaNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func collectionExists(ctx context.Context, hdb *fsql.DB, schema, table string) (bool, error) {
	sql := fmt.Sprintf("SELECT count(*) FROM M_TABLES "+
		"WHERE TABLE_TYPE = 'COLLECTION' AND "+
		"SCHEMA_NAME = '%s' AND TABLE_NAME = '%s'",
		schema, table,
	)

	var count int
	if err := hdb.QueryRowContext(ctx, sql).Scan(&count); err != nil {
		return false, lazyerrors.Error(err)
	}

	return (count == 1), nil
}

// CreateCollection creates a new SAP HANA JSON Document Store collection.
// Schema will automatically be created if it does not exist.
//
// It returns ErrAlreadyExist if collection already exist.
func createCollection(ctx context.Context, hdb *fsql.DB, schema, table string) error {
	err := createSchemaIfNotExists(ctx, hdb, schema)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("CREATE COLLECTION %q.%q", schema, table)

	_, err = hdb.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

// CreateCollectionIfNotExists creates a new SAP HANA JSON Document Store collection.
//
// Returns nil if collection already exist.
func createCollectionIfNotExists(ctx context.Context, hdb *fsql.DB, schema, table string) error {
	exists, err := collectionExists(ctx, hdb, schema, table)
	if err != nil {
		return getHanaErrorIfExists(err)
	}

	if !exists {
		err = createCollection(ctx, hdb, schema, table)
	}

	return err
}

func dropCollection(ctx context.Context, hdb *fsql.DB, schema, table string) (bool, error) {
	sql := fmt.Sprintf("DROP COLLECTION %q.%q CASCADE", schema, table)

	_, err := hdb.ExecContext(ctx, sql)
	if err != nil {
		if hanaErrorCollectionNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
