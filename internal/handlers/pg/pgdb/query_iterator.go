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

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type queryRowsIterator interface {
	Close()
	Next() bool
	Err() error
	Scan(...interface{}) error
}

type queryIterator struct {
	logger  pgx.Logger
	rows    queryRowsIterator
	ctx     context.Context
	sp      SQLParam
	hasNext bool
	started bool
}

func (qi *queryIterator) Close() {
	if qi.rows != nil {
		qi.rows.Close()
	}
}

// Next returns true if exists result in the query.
func (qi *queryIterator) Next() bool {
	if qi.rows == nil {
		return false
	}

	if qi.started {
		return qi.hasNext
	}

	qi.started = true
	qi.hasNext = qi.rows.Next()

	return qi.hasNext
}

// DocumentsFiltered implementing a query iterator to filter documents.
//
// The number of documents returned is defined by QueryIteratorSliceCapacity.
//
// Context cancellation is not considered an error.
//
// Need to call Next() before fetching and filtering results.
func (qi *queryIterator) DocumentsFiltered(filter *types.Document) ([]*types.Document, error) {
	res := make([]*types.Document, 0, QueryIteratorSliceCapacity)
	if qi.rows == nil {
		return res, nil
	}

	for i := 0; i < QueryIteratorSliceCapacity; i++ {
		if !qi.hasNext {
			break
		}

		if err := qi.ctx.Err(); err != nil {
			info := map[string]any{
				"db":         qi.sp.DB,
				"collection": qi.sp.Collection,
				"error":      err,
			}

			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				qi.logger.Log(qi.ctx, pgx.LogLevelWarn, "context canceled, stopping fetching", info)
				break
			}

			qi.logger.Log(qi.ctx, pgx.LogLevelError, "got error, stopping fetching", info)

			return nil, err
		}

		doc, match, err := qi.scanFilter(filter)
		if err != nil {
			return nil, err
		}

		if match {
			res = append(res, doc)
		}

		qi.hasNext = qi.rows.Next()
	}

	if qi.rows.Err() != nil {
		return nil, qi.rows.Err()
	}

	return res, nil
}

func (qi *queryIterator) scanFilter(filter *types.Document) (*types.Document, bool, error) {
	var b []byte
	if err := qi.rows.Scan(&b); err != nil {
		return nil, false, lazyerrors.Error(err)
	}

	data, err := fjson.Unmarshal(b)
	if err != nil {
		return nil, false, lazyerrors.Error(err)
	}

	doc := data.(*types.Document)

	match, err := common.FilterDocument(doc, filter)
	if err != nil {
		return nil, false, err
	}

	return doc, match, nil
}

func newQueryIterator(ctx context.Context, logger pgx.Logger, rows pgx.Rows, params SQLParam) types.DocumentIterator {
	return &queryIterator{
		rows:   rows,
		ctx:    ctx,
		logger: logger,
		sp:     params,
	}
}
