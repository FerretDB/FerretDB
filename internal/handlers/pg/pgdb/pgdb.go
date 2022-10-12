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

// Package pgdb provides PostgreSQL connection utilities.
package pgdb

import "fmt"

// Errors are wrapped with lazyerrors.Error,
// so the caller needs to use errors.Is to check the error,
// for example, errors.Is(err, ErrSchemaNotExist).
var (
	// ErrTableNotExist indicates that there is no such table.
	ErrTableNotExist = fmt.Errorf("collection/table does not exist")

	// ErrSchemaNotExist indicates that there is no such schema.
	ErrSchemaNotExist = fmt.Errorf("database/schema does not exist")

	// ErrAlreadyExist indicates that a schema or table already exists.
	ErrAlreadyExist = fmt.Errorf("database/schema or collection/table already exist")

	// ErrInvalidTableName indicates that a schema or table didn't passed name checks.
	ErrInvalidTableName = fmt.Errorf("invalid collection/table name")

	// ErrInvalidDatabaseName indicates that a database name didn't passed checks.
	ErrInvalidDatabaseName = fmt.Errorf("invalid database/schema name")
)
