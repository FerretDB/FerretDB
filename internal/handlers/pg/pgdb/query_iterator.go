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
	"errors"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Iterator implements iterator.Interface to fetch documents from the database.
type Iterator struct {
	ctx  context.Context
	rows pgx.Rows
}

// NewIterator returns a new iterator for the given SQL param.
// If the table for the given SQL params doesn't exist, it returns nil and no error.
// If an error occurs, it returns nil and "lazy error".
func NewIterator(ctx context.Context, tx pgx.Tx, sp *SQLParam) (*Iterator, error) {
	q, args, err := buildQuery(ctx, tx, sp)
	if err != nil {
		if errors.Is(err, ErrTableNotExist) {
			return nil, nil
		}

		return nil, lazyerrors.Error(err)
	}

	rows, err := tx.Query(ctx, q, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Iterator{
		ctx:  ctx,
		rows: rows,
	}, nil
}

// Next implements iterator.Interface.
func (it *Iterator) Next() (*types.Document, error) {
	if err := it.ctx.Err(); err != nil {
		return nil, err
	}

	if !it.rows.Next() {
		return nil, iterator.ErrIteratorDone
	}

	var b []byte
	if err := it.rows.Scan(&b); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := pjson.Unmarshal(b)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc.(*types.Document), nil
}
