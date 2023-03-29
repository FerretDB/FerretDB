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

package pgdb

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// queryIterator implements iterator.Interface to fetch documents from the database.
type queryIterator struct {
	ctx       context.Context
	unmarshal func(b []byte) (*types.Document, error) // defaults to pjson.Unmarshal

	m    sync.Mutex
	rows pgx.Rows

	token *resource.Token
}

// newIterator returns a new queryIterator for the given pgx.Rows.
//
// Iterator's Close method closes rows.
//
// Nil rows are possible and return already done iterator.
func newIterator(ctx context.Context, rows pgx.Rows, p *iteratorParams) types.DocumentsIterator {
	unmarshalFunc := p.unmarshal
	if unmarshalFunc == nil {
		unmarshalFunc = pjson.Unmarshal
	}

	iter := &queryIterator{
		ctx:       ctx,
		unmarshal: unmarshalFunc,
		rows:      rows,
		token:     resource.NewToken(),
	}

	resource.Track(iter, iter.token)

	return iter
}

// Next implements iterator.Interface.
//
// Errors (possibly wrapped) are:
//   - iterator.ErrIteratorDone;
//   - context.Canceled;
//   - context.DeadlineExceeded;
//   - something else.
//
// Otherwise, as the first value it returns the number of the current iteration (starting from 0),
// as the second value it returns the document.
func (iter *queryIterator) Next() (struct{}, *types.Document, error) {
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
		if err := iter.rows.Err(); err != nil {
			return unused, nil, lazyerrors.Error(err)
		}

		// to avoid context cancellation changing the next `Next()` error
		// from `iterator.ErrIteratorDone` to `context.Canceled`
		iter.close()

		return unused, nil, iterator.ErrIteratorDone
	}

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
	iter.m.Lock()
	defer iter.m.Unlock()

	iter.close()
}

// close closes iterator without holding mutex.
//
// This should be called only when the caller already holds the mutex.
func (iter *queryIterator) close() {
	if iter.rows != nil {
		iter.rows.Close()
		iter.rows = nil
	}

	resource.Untrack(iter, iter.token)
}

// check interfaces
var (
	_ types.DocumentsIterator = (*queryIterator)(nil)
)
