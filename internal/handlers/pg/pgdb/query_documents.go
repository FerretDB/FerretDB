package pgdb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/jackc/pgx/v4"
)

const (
	fetchedChannelCapacity = 3
	fetchedSliceCapacity   = 2
)

type FetchedDocs struct {
	Docs []*types.Document
	Err  error
}

// QueryDocuments returns a list of documents for given FerretDB database and collection.
func (pgPool *Pool) QueryDocuments(ctx context.Context, db, collection, comment string) (<-chan FetchedDocs, error) {
	fetchedChan := make(chan FetchedDocs, fetchedChannelCapacity)

	// errBeforeFetching indicates that an error occurred before fetching started.
	var errBeforeFetching error

	// waitFetching signals when fetching is started,
	// if errors occur before fetching is started, they are returned immediately.
	var waitFetching = make(chan struct{})

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

		res := make([]*types.Document, 0, fetchedSliceCapacity)
		for i := 0; i < fetchedSliceCapacity; i++ {
			if !rows.Next() {
				return nil
			}

			var b []byte
			if err := rows.Scan(&b); err != nil {
				return writeFetched(ctx, fetched, FetchedDocs{Err: lazyerrors.Error(err)})
			}

			doc, err := fjson.Unmarshal(b)
			if err != nil {
				return writeFetched(ctx, fetched, FetchedDocs{Err: lazyerrors.Error(err)})
			}

			res = append(res, doc.(*types.Document))
		}

		if err := writeFetched(ctx, fetched, FetchedDocs{Docs: res}); err != nil {
			return err
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
