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
// for example, errors.Is(err, ErrDatabaseNotExist).
var (
	// ErrTableNotExist indicates that there is no such table.
	ErrTableNotExist = fmt.Errorf("collection/table does not exist")

	// ErrDatabaseNotExist indicates that there is no such Database.
	ErrDatabaseNotExist = fmt.Errorf("database/Database does not exist")

	// ErrDatabaseAlreadyExist indicates that a Database already exists.
	ErrDatabaseAlreadyExist = fmt.Errorf("database/Database already exists")

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
	362: ErrDatabaseNotExist,
	386: ErrDatabaseAlreadyExist,
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

// hanaErrorDatabaseNotExist checks if the error is about collection not existing.
//
// Returns true if the err is database does not exist.
func hanaErrorDatabaseNotExist(err error) bool {
	var dbError driver.Error
	if errors.As(err, &dbError) {
		return dbError.Code() == 362
	}

	return false
}

func databaseExists(ctx context.Context, hdb *fsql.DB, database string) (bool, error) {
	sql := fmt.Sprintf("SELECT COUNT(*) FROM \"PUBLIC\".\"SCHEMAS\" WHERE SCHEMA_NAME = '%s'", database)

	var count int
	if err := hdb.QueryRowContext(ctx, sql).Scan(&count); err != nil {
		return false, lazyerrors.Error(err)
	}

	return count == 1, nil
}

// CreateDatabase creates a database in SAP HANA JSON Document Store.
//
// Returns ErrDatabaseAlreadyExist if database already exists.
func createDatabase(ctx context.Context, hdb *fsql.DB, database string) error {
	sqlStmt := fmt.Sprintf("CREATE SCHEMA %q", database)

	_, err := hdb.ExecContext(ctx, sqlStmt)

	return getHanaErrorIfExists(err)
}

func createDatabaseIfNotExists(ctx context.Context, hdb *fsql.DB, database string) error {
	exists, err := databaseExists(ctx, hdb, database)
	if err != nil {
		return getHanaErrorIfExists(err)
	}

	if !exists {
		err = createDatabase(ctx, hdb, database)
	}

	return err
}

// DropDatabase drops database.
//
// Returns ErrDatabaseNotExist if Database does not exist.
func dropDatabase(ctx context.Context, hdb *fsql.DB, database string) (bool, error) {
	sql := fmt.Sprintf("DROP SCHEMA %q CASCADE", database)

	_, err := hdb.ExecContext(ctx, sql)
	if err != nil {
		if hanaErrorDatabaseNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func collectionExists(ctx context.Context, hdb *fsql.DB, database, table string) (bool, error) {
	sql := fmt.Sprintf("SELECT count(*) FROM M_TABLES "+
		"WHERE TABLE_TYPE = 'COLLECTION' AND "+
		"SCHEMA_NAME = '%s' AND TABLE_NAME = '%s'",
		database, table,
	)

	var count int
	if err := hdb.QueryRowContext(ctx, sql).Scan(&count); err != nil {
		return false, lazyerrors.Error(err)
	}

	return (count == 1), nil
}

// CreateCollection creates a new SAP HANA JSON Document Store collection.
// Database will automatically be created if it does not exist.
//
// It returns ErrAlreadyExist if collection already exist.
func createCollection(ctx context.Context, hdb *fsql.DB, database, table string) error {
	err := createDatabaseIfNotExists(ctx, hdb, database)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf("CREATE COLLECTION %q.%q", database, table)

	_, err = hdb.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

// CreateCollectionIfNotExists creates a new SAP HANA JSON Document Store collection.
//
// Returns nil if collection already exist.
func createCollectionIfNotExists(ctx context.Context, hdb *fsql.DB, database, table string) error {
	exists, err := collectionExists(ctx, hdb, database, table)
	if err != nil {
		return getHanaErrorIfExists(err)
	}

	if !exists {
		err = createCollection(ctx, hdb, database, table)
	}

	return err
}

func dropCollection(ctx context.Context, hdb *fsql.DB, database, table string) (bool, error) {
	sql := fmt.Sprintf("DROP COLLECTION %q.%q CASCADE", database, table)

	_, err := hdb.ExecContext(ctx, sql)
	if err != nil {
		if hanaErrorCollectionNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
