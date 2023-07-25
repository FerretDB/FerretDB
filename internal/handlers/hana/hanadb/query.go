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
	"database/sql"
	"sync"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// QueryParams represents options/parameters used for SQL query/statement.
type QueryParams struct {
	DB         string
	Collection string
}

// queryIterator implements iterator.Interface to fetch documents from the database.
type queryIterator struct {
	ctx       context.Context
	unmarshal func(b []byte) (*types.Document, error) // defaults to sjson.Unmarshal

	m    sync.Mutex
	rows *sql.Rows

	token *resource.Token
}

// QueryDocuments returns an queryIterator to fetch documents for given SQLParams.
// If the collection doesn't exist, it returns an empty iterator and no error.
// If an error occurs, it returns nil and that error, possibly wrapped.
func (hanaPool *Pool) QueryDocuments(ctx context.Context, qp *QueryParams) (types.DocumentsIterator, error) {
	// Todo: build correct SQL here
	sqlStmt := "SELECT $1 FROM $1"

	rows, err := hanaPool.QueryContext(ctx, sqlStmt, qp.Collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	iter := &queryIterator{
		ctx:       ctx,
		unmarshal: sjson.Unmarshal,
		rows:      rows,
	}

	return iter, nil
}

// Next implements iterator.Interface.
//
// Otherwise, the next document is returned.
func (iter *queryIterator) Next() (struct{}, *types.Document, error) {
	defer observability.FuncCall(iter.ctx)()

	iter.m.Lock()
	defer iter.m.Unlock()

	var unused struct{}

	var b []byte
	if err := iter.rows.Scan(&b); err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	doc, err := iter.unmarshal(b)
	if err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	return unused, doc, nil
}

// Close implements iterator.Interface.
func (iter *queryIterator) Close() {
	defer observability.FuncCall(iter.ctx)()

	iter.m.Lock()
	defer iter.m.Unlock()

	iter.close()
}

// close closes iterator without holding mutex.
//
// This should be called only when the caller already holds the mutex.
func (iter *queryIterator) close() {
	defer observability.FuncCall(iter.ctx)()

	if iter.rows != nil {
		iter.rows.Close()
		iter.rows = nil
	}

	resource.Untrack(iter, iter.token)
}
