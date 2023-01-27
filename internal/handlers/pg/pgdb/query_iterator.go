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
	"sync/atomic"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// queryIterator implements iterator.Interface to fetch documents from the database.
type queryIterator struct {
	ctx   context.Context
	n     atomic.Uint32
	rows  pgx.Rows
	stack []byte
}

// newIterator returns a new queryIterator for the given pgx.Rows.
//
// Iterator's Close method closes rows.
//
// Nil rows are possible and return already done iterator.
func newIterator(ctx context.Context, rows pgx.Rows) iterator.Interface[int, *types.Document] {
	iter := &queryIterator{
		ctx:   ctx,
		rows:  rows,
		stack: debugbuild.Stack(),
	}

	runtime.SetFinalizer(iter, func(iter *queryIterator) {
		panic("queryIterator.Close() has not been called:\n" + string(iter.stack))
	})

	return iter
}

// Next implements iterator.Interface.
//
// Errors (possibly wrapped) are:
//   - context.Canceled;
//   - context.DeadlineExceeded;
//   - iterator.ErrIteratorDone;
//   - something else.
//
// Otherwise, as the first value it returns the number of the current iteration (starting from 0),
// as the second value it returns the document.
func (iter *queryIterator) Next() (int, *types.Document, error) {
	if iter.rows == nil || !iter.rows.Next() {
		return 0, nil, iterator.ErrIteratorDone
	}

	if err := iter.ctx.Err(); err != nil {
		return 0, nil, err
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

	return int(n), doc, nil
}

// Close implements iterator.Interface.
func (iter *queryIterator) Close() {
	if iter.rows != nil {
		iter.rows.Close()
	}

	runtime.SetFinalizer(iter, nil)
}

// check interfaces
var (
	_ iterator.Interface[int, *types.Document] = (*queryIterator)(nil)
)
