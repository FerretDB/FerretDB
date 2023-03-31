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
	ErrAlreadyExist = fmt.Errorf("database/schema or collection/table already exists")

	// ErrIndexKeyAlreadyExist indicates that an index key already exists with a different name.
	ErrIndexKeyAlreadyExist = fmt.Errorf("index key already exists with a different name")

	// ErrIndexNameAlreadyExist indicates that an index name already exists with a different key.
	ErrIndexNameAlreadyExist = fmt.Errorf("index name already exists with a different key")

	// ErrIndexNotExist indicates there is no such index.
	ErrIndexNotExist = fmt.Errorf("index does not exist")

	// ErrIndexCannotDelete indicates the index cannot be deleted.
	ErrIndexCannotDelete = fmt.Errorf("index cannot be deleted")

	// ErrInvalidCollectionName indicates that a collection didn't pass name checks.
	ErrInvalidCollectionName = fmt.Errorf("invalid FerretDB collection name")

	// ErrInvalidDatabaseName indicates that a database didn't pass name checks.
	ErrInvalidDatabaseName = fmt.Errorf("invalid FerretDB database name")

	// ErrUniqueViolation indicates that operations violates a unique constraint.
	ErrUniqueViolation = fmt.Errorf("unique constraint violation")
)
