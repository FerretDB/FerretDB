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
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/SAP/go-hdb/driver"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
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
	rows      *sql.Rows
	token     *resource.Token
	unmarshal func(data []byte) (*types.Document, error)
	ctx       context.Context
	m         sync.Mutex
}

// QueryDocuments returns an queryIterator to fetch documents for given SQLParams.
// If the collection doesn't exist, it returns an empty iterator and no error.
// If an error occurs, it returns nil and that error, possibly wrapped.
func (hanaPool *Pool) QueryDocuments(ctx context.Context, qp *QueryParams) (types.DocumentsIterator, error) {
	// Todo: build correct SQL here

	sqlStmt := fmt.Sprintf("SELECT * FROM %q.%q", qp.DB, qp.Collection)
	rows, err := hanaPool.QueryContext(ctx, sqlStmt)

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	iter := &queryIterator{
		ctx:       ctx,
		unmarshal: unmarshal,
		rows:      rows,
		token:     resource.NewToken(),
	}
	resource.Track(iter, iter.token)

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
	// ignore context error, if any, if iterator is already closed
	if iter.rows == nil {
		return unused, nil, iterator.ErrIteratorDone
	}

	if err := context.Cause(iter.ctx); err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	if !iter.rows.Next() {
		err := iter.rows.Err()

		iter.close()

		if err == nil {
			err = iterator.ErrIteratorDone
		}

		return unused, nil, lazyerrors.Error(err)
	}

	b := new(bytes.Buffer)
	lob := new(driver.Lob)
	lob.SetWriter(b)

	if err := iter.rows.Scan(lob); err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	doc, err := iter.unmarshal(b.Bytes())
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
