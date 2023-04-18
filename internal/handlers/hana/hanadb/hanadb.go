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

// Package hanadb provides SAP HANA connection utilities.
package hanadb

import (
	"errors"
	"fmt"

	"github.com/SAP/go-hdb/driver"
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
// Returns one of the errors above or the original error if the formatted version doesn't exist
func getHanaErrorIfExists(err error) error {
	var dbError driver.Error
	if errors.As(err, &dbError) {
		if hanaErr, ok := Errors[dbError.Code()]; ok {
			return hanaErr
		}
	}

	return err
}
