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

// CreateCollection creates a new SAP HANA JSON Document Store collection.
//
// It returns ErrAlreadyExist if collection already exist.
func (hanaPool *Pool) CreateCollection(ctx context.Context, qp *QueryParams) error {
	sql := fmt.Sprintf("CREATE COLLECTION %q.%q", qp.DB, qp.Collection)

	_, err := hanaPool.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

// CreateCollectionIfNotExists creates a new SAP HANA JSON Document Store collection.
//
// Returns nil if collection already exist.
func (hanaPool *Pool) CreateCollectionIfNotExists(ctx context.Context, qp *QueryParams) error {
	err := hanaPool.CreateCollection(ctx, qp)

	switch {
	case errors.Is(err, ErrCollectionAlreadyExist):
		return nil
	default:
		return err
	}
}

// DropCollection drops collection.
//
// Returns ErrTableNotExist is collection does not exist.
func (hanaPool *Pool) DropCollection(ctx context.Context, qp *QueryParams) error {
	sql := fmt.Sprintf("DROP COLLECTION %q.%q", qp.DB, qp.Collection)

	_, err := hanaPool.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

// CollectionSize calculates the size of the given collection.
//
// Returns 0 if schema or collection doesn't exist, otherwise returns its size.
func (hanaPool *Pool) CollectionSize(ctx context.Context, qp *QueryParams) (int64, error) {
	sqlStmt := "SELECT TABLE_SIZE FROM M_TABLES WHERE SCHEMA_NAME = $1 AND TABLE_NAME = $2 AND TABLE_TYPE = 'COLLECTION'"

	var size any
	if err := hanaPool.QueryRowContext(ctx, sqlStmt, qp.DB, qp.Collection).Scan(&size); err != nil {
		return 0, lazyerrors.Error(err)
	}

	var collectionSize int64
	switch size := size.(type) {
	case int64:
		collectionSize = size
	case nil:
		collectionSize = 0
	default:
		return 0, lazyerrors.Errorf("Got wrong type for tableSize. Got: %T", collectionSize)
	}

	return collectionSize, nil
}
