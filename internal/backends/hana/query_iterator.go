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

package hana

import (
	"bytes"
	"context"
	"sync"

	"github.com/SAP/go-hdb/driver"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// queryIterator implements iterator.Interface to fetch documents from the database.
type queryIterator struct {
	rows  *fsql.Rows
	token *resource.Token
	ctx   context.Context
	m     sync.Mutex
}

func newQueryIterator(ctx context.Context, rows *fsql.Rows) types.DocumentsIterator {
	iter := &queryIterator{
		rows:  rows,
		token: resource.NewToken(),
		ctx:   ctx,
	}
	resource.Track(iter, iter.token)

	return iter
}

// Next implements iterator.Interface.
// Otherwise, the next document is returned.
func (iter *queryIterator) Next() (struct{}, *types.Document, error) {
	_, cancel := observability.FuncCall(iter.ctx)
	defer cancel(nil)

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

	doc, err := unmarshalHana(b.Bytes())
	if err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	return unused, doc, nil
}

// Close implements iterator.Interface.
func (iter *queryIterator) Close() {
	_, cancel := observability.FuncCall(iter.ctx)
	defer cancel(nil)

	iter.m.Lock()
	defer iter.m.Unlock()

	iter.close()
}

// close closes iterator without holding mutex.
//
// This should be called only when the caller already holds the mutex.
func (iter *queryIterator) close() {
	_, cancel := observability.FuncCall(iter.ctx)
	defer cancel(nil)

	if iter.rows != nil {
		iter.rows.Close()
		iter.rows = nil
	}

	resource.Untrack(iter, iter.token)
}
