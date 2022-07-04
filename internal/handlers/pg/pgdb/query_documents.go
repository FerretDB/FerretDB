package pgdb

import (
	"context"
	"log"
	"strings"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
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

	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	table, err := pgPool.getTableName(ctx, tx, db, collection)
	if err != nil {
		close(fetchedChan)
		return fetchedChan, err
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
		close(fetchedChan)
		return fetchedChan, lazyerrors.Error(err)
	}

	go func() {

		/// ???? What to do with transaction? With channels it will hang for a lot of time
		defer func() {
			if err != nil {
				pgPool.logger.Error("failed to perform rollback", zap.Error(tx.Rollback(ctx)))
				return
			}
			pgPool.logger.Error("failed to perform commit", zap.Error(tx.Commit(ctx)))
		}()

		defer close(fetchedChan)
		defer rows.Close()

		var ctxCanceled bool
		defer func(canceled bool) {
			if canceled {
				pgPool.logger.Info("got a signal to stop fetching, fetch canceled",
					zap.String("db", db), zap.String("collection", collection),
				)
			}
		}(ctxCanceled)

		for {
			select {
			case <-ctx.Done():
				ctxCanceled = true
				return
			default:
				// fetch next batch of documents
			}

			res := make([]*types.Document, 0, fetchedSliceCapacity)
			for i := 0; i < len(res); i++ {
				if !rows.Next() {
					break
				}

				var b []byte
				if err := rows.Scan(&b); err != nil {
					ctxCanceled = !writeFetched(ctx, fetchedChan, FetchedDocs{Err: lazyerrors.Error(err)})
					return
				}

				doc, err := fjson.Unmarshal(b)
				if err != nil {
					ctxCanceled = !writeFetched(ctx, fetchedChan, FetchedDocs{Err: lazyerrors.Error(err)})
					return
				}

				res = append(res, doc.(*types.Document))
				log.Fatal(doc)
			}

			if ctxCanceled = !writeFetched(ctx, fetchedChan, FetchedDocs{Docs: res}); ctxCanceled {
				return
			}
		}
	}()

	return fetchedChan, nil
}

// writeFetched sends `FetchedDocs` to `fetched` channel or handles context cancellation.
// It returns `true` if `FetchedDocs` was sent successfully or `false` if context cancellation was received.
func writeFetched(ctx context.Context, fetched chan FetchedDocs, doc FetchedDocs) bool {
	select {
	case <-ctx.Done():
		return false
	case fetched <- doc:
		return true
	}
}
