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
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// queryIterator implements iterator.Interface to fetch documents from the database.
type queryIterator struct {
	ctx       context.Context
	rows      pgx.Rows
	n         atomic.Uint32
	closeOnce sync.Once
}

// newIterator returns a new queryIterator for the given pgx.Rows.
// It sets finalizer to close the rows.
func newIterator(ctx context.Context, rows pgx.Rows) iterator.Interface[uint32, *types.Document] {
	// queryIterator is defined as pointer to address iter in the finalizer.
	iter := &queryIterator{
		ctx:  ctx,
		rows: rows,
	}

	runtime.SetFinalizer(iter, func(it *queryIterator) {
		it.closeOnce.Do(func() {
			panic("queryIterator.Close() has not been called")
		})
	})

	return iter
}

// Next implements iterator.Interface.
//
// If an error occurs, it returns 0, nil, and the error.
// Possible errors are: context.Canceled, context.DeadlineExceeded, and lazy error.
// Otherwise, as the first value it returns the number of the current iteration (starting from 0),
// as the second value it returns the document.
func (iter *queryIterator) Next() (uint32, *types.Document, error) {
	if err := iter.ctx.Err(); err != nil {
		return 0, nil, err
	}

	if iter.rows == nil || !iter.rows.Next() {
		return 0, nil, iterator.ErrIteratorDone
	}

	var b []byte
	if err := iter.rows.Scan(&b); err != nil {
		return 0, nil, lazyerrors.Error(err)
	}

	doc, err := pjson.Unmarshal(b)
	if err != nil {
		return 0, nil, lazyerrors.Error(err)
	}

	n := iter.n.Add(1) - 1

	return n, doc, nil
}

// Close implements iterator.Interface.
func (iter *queryIterator) Close() {
	iter.closeOnce.Do(func() {
		if iter.rows != nil {
			iter.rows.Close()
		}
	})
}

// check interfaces
var (
	_ iterator.Interface[uint32, *types.Document] = (*queryIterator)(nil)
)
