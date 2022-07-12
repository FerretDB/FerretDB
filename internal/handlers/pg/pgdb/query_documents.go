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
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

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

// QueryDocuments returns a list of documents for given FerretDB database and collection.
func (pgPool *Pool) QueryDocuments(ctx context.Context, db, collection, comment string) (<-chan FetchedDocs, error) {
	fetchedChan := make(chan FetchedDocs, FetchedChannelBufSize)

	// errBeforeFetching indicates that an error occurred before fetching started.
	var errBeforeFetching error

	// waitFetching signals when fetching is started,
	// if errors occur before fetching is started, they are returned immediately.
	waitFetching := make(chan struct{})

	go func() {
		defer close(fetchedChan)

		err := pgPool.inTransaction(ctx, func(tx pgx.Tx) error {
			table, err := pgPool.getTableName(ctx, tx, db, collection)
			if err != nil {
				errBeforeFetching = err
				close(waitFetching)
				return err
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
				errBeforeFetching = err
				close(waitFetching)
				return lazyerrors.Error(err)
			}
			defer rows.Close()

			close(waitFetching)
			return iterateFetch(ctx, fetchedChan, rows)
		})

		switch {
		case err == nil:
			// nothing
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			pgPool.logger.Warn(
				fmt.Sprintf("caught %v, stop fetching", err),
				zap.String("db", db), zap.String("collection", collection),
			)
		default:
			pgPool.logger.Error("exiting fetching with an error", zap.Error(err))
		}
	}()

	<-waitFetching
	return fetchedChan, errBeforeFetching
}

func iterateFetch(ctx context.Context, fetched chan FetchedDocs, rows pgx.Rows) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// fetch next batch of documents
		}

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

		if len(res) > 0 {
			if err := writeFetched(ctx, fetched, FetchedDocs{Docs: res}); err != nil {
				return err
			}
		}

		if allFetched {
			return nil
		}
	}
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
