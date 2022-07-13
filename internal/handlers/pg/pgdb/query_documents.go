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
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

const (
	// FetchedChannelBufSize is the size of the buffer of the channel that is used in QueryDocuments.
	FetchedChannelBufSize = 3
	// FetchedSliceCapacity is the capacity of the slice in FetchedDocs.
	FetchedSliceCapacity = 2
)

// FetchedDocs is a struct that contains a list of documents and an error.
// It is used in the fetched channel returned by QueryDocuments.
type FetchedDocs struct {
	Docs []*types.Document
	Err  error
}

// QueryDocuments returns a channel with buffer FetchedChannelBufSize
// to fetch list of documents for given FerretDB database and collection.
//
// If an error occurs before fetching started, it returns a closed channel and an error.
//
// Fetched documents are sent to the channel as well as errors.
// The channel is closed when the query is finished.
// The channel is also closed if an error occurs or context cancellation is received.
func (pgPool *Pool) QueryDocuments(ctx context.Context, db, collection, comment string) (<-chan FetchedDocs, error) {
	fetchedChan := make(chan FetchedDocs, FetchedChannelBufSize)

	tx, err := pgPool.Begin(ctx)
	if err != nil {
		close(fetchedChan)
		return fetchedChan, lazyerrors.Error(err)
	}

	table, err := pgPool.getTableName(ctx, tx, db, collection)
	if err != nil {
		return fetchedChan, lazyerrors.Error(err)
	}

	sql := `SELECT _jsonb `
	if comment != "" {
		comment = strings.ReplaceAll(comment, "/*", "/ *")
		comment = strings.ReplaceAll(comment, "*/", "* /")

		sql += `/* ` + comment + ` */ `
	}
	sql += `FROM ` + pgx.Identifier{db, table}.Sanitize()

	rows, err := tx.Query(ctx, sql)
	if err != nil {
		_ = tx.Rollback(ctx)
		close(fetchedChan)
		return fetchedChan, lazyerrors.Error(err)
	}

	go func() {
		defer close(fetchedChan)

		_ = iterateFetch(ctx, fetchedChan, rows)

		_ = tx.Rollback(ctx)
	}()

	/*switch {
	case err == nil:
		// nothing
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		pgPool.logger.Warn(
			fmt.Sprintf("caught %v, stop fetching", err),
			zap.String("db", db), zap.String("collection", collection),
		)
	default:
		pgPool.logger.Error("exiting fetching with an error", zap.Error(err))
	}*/

	return fetchedChan, nil
}

// iterateFetch iterates over the rows returned by the query and sends FetchedDocs to fetched channel.
// It returns ctx.Err() if context cancellation was received.
func iterateFetch(ctx context.Context, fetched chan FetchedDocs, rows pgx.Rows) error {
	for ctx.Err() == nil {
		var allFetched bool
		res := make([]*types.Document, 0, FetchedSliceCapacity)
		for i := 0; i < FetchedSliceCapacity; i++ {
			if !rows.Next() {
				allFetched = true
				break
			}

			var b []byte
			if err := rows.Scan(&b); err != nil {
				// TODO: cover this case with a test
				return writeFetched(ctx, fetched, FetchedDocs{Err: lazyerrors.Error(err)})
			}

			doc, err := fjson.Unmarshal(b)
			if err != nil {
				// TODO: cover this case with a test
				return writeFetched(ctx, fetched, FetchedDocs{Err: lazyerrors.Error(err)})
			}

			res = append(res, doc.(*types.Document))
		}

		if err := rows.Err(); err != nil {
			panic(err) // TODO
		}

		if len(res) > 0 {
			if err := writeFetched(ctx, fetched, FetchedDocs{Docs: res}); err != nil {
				return err
			}
		}

		if allFetched {
			return nil
		}
	}

	return ctx.Err()
}

// writeFetched sends FetchedDocs to fetched channel or handles context cancellation.
// It returns ctx.Err() if context cancellation was received.
func writeFetched(ctx context.Context, fetched chan FetchedDocs, doc FetchedDocs) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case fetched <- doc:
		return nil
	}
}
