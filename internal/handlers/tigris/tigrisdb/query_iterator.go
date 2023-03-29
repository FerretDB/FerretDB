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

package tigrisdb

import (
	"context"
	"sync"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// queryIterator implements iterator.Interface to fetch documents from the database.
type queryIterator struct {
	ctx    context.Context
	schema *tjson.Schema

	m    sync.Mutex
	iter driver.Iterator

	token *resource.Token
}

// newIterator returns a new queryIterator for the given driver.Iterator.
//
// Iterator's Close method closes driver.Iterator.
//
// No documents are possible and return already done iterator.
func newQueryIterator(ctx context.Context, titer driver.Iterator, schema *tjson.Schema) types.DocumentsIterator {
	iter := &queryIterator{
		ctx:    ctx,
		schema: schema,
		iter:   titer,
		token:  resource.NewToken(),
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
	if iter.iter == nil {
		return unused, nil, iterator.ErrIteratorDone
	}

	if err := context.Cause(iter.ctx); err != nil {
		return unused, nil, err
	}

	var document driver.Document

	ok := iter.iter.Next(&document)
	if !ok {
		err := iter.iter.Err()

		switch {
		case err == nil:
			// nothing
		case IsInvalidArgument(err):
			// Skip errors from filtering different types.
			// For example, given document {v: 42} and filter {v: "42"},
			// MongoDB would skip that document because the type is different.
			// Tigris returns a schema error in such cases that we ignore.
		default:
			return unused, nil, lazyerrors.Error(err)
		}

		// to avoid context cancellation changing the next `Next()` error
		// from `iterator.ErrIteratorDone` to `context.Canceled`
		iter.close()

		return unused, nil, iterator.ErrIteratorDone
	}

	doc, err := tjson.Unmarshal(document, iter.schema)
	if err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	return unused, doc.(*types.Document), nil
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
	if iter.iter != nil {
		iter.iter.Close()

		iter.iter = nil
	}

	resource.Untrack(iter, iter.token)
}

// check interfaces
var (
	_ types.DocumentsIterator = (*queryIterator)(nil)
)
