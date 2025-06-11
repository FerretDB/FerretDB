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

package documentdb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/FerretDB/wire/wirebson"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// GetMore returns the next page of the cursor.
// It is a part of the implementation of the `getMore` command.
func (p *Pool) GetMore(ctx context.Context, db string, spec wirebson.RawDocument, cursorID int64) (wirebson.RawDocument, error) {
	ctx, span := otel.Tracer("").Start(ctx, "pool.GetMore")
	defer span.End()

	continuation, conn := p.r.GetCursor(cursorID)
	if continuation == nil {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrCursorNotFound,
			fmt.Sprintf("cursor id %d not found", cursorID),
			"getMore",
		)
	}

	if conn == nil {
		poolConn, err := p.Acquire()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		defer poolConn.Release()

		conn = poolConn.Conn()
	}

	page, continuation, err := documentdb_api.CursorGetMore(ctx, conn, p.l, db, spec, continuation)
	if err != nil {
		p.r.CloseCursor(ctx, cursorID)
		return nil, lazyerrors.Error(err)
	}

	p.l.DebugContext(
		ctx, "GetMore result", slog.Int64("cursor", cursorID),
		slog.Any("page", page), slog.Any("continuation", continuation),
	)

	p.r.UpdateCursor(cursorID, continuation)

	return page, nil
}

// KillCursor closes the cursor with the given id and removes it from the registry.
// It returns true if the cursor was found and removed.
// It is a part of the implementation of the `killCursors` command.
//
// It attempts a clean close by sending the exit message to PostgreSQL.
// However, this could block so ctx is available to limit the time to wait (up to 3 seconds).
// The underlying connection will always be called regardless of any other errors.
func (p *Pool) KillCursor(ctx context.Context, id int64) bool {
	ctx, span := otel.Tracer("").Start(ctx, "pool.KillCursor")
	defer span.End()

	return p.r.CloseCursor(ctx, id)
}

// ListCollections returns the first page of the `listCollections` cursor and the cursor ID.
func (p *Pool) ListCollections(ctx context.Context, db string, spec wirebson.RawDocument) (wirebson.RawDocument, int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "pool.ListCollections")
	defer span.End()

	poolConn, err := p.Acquire()
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}
	defer poolConn.Release()

	conn := poolConn.Conn()

	page, continuation, persist, cursorID, err := documentdb_api.ListCollectionsCursorFirstPage(ctx, conn, p.l, db, spec, 0)
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}

	p.l.DebugContext(
		ctx, "ListCollections result",
		slog.Any("page", page), slog.Any("continuation", continuation),
		slog.Bool("persist", persist), slog.Int64("cursor", cursorID),
	)

	if persist {
		conn = poolConn.hijack()
	} else {
		conn = nil
	}

	p.r.NewCursor(cursorID, continuation, conn)

	return page, cursorID, nil
}

// Find returns the first page of the `find` cursor and the cursor ID.
func (p *Pool) Find(ctx context.Context, db string, spec wirebson.RawDocument) (wirebson.RawDocument, int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "pool.Find")
	defer span.End()

	poolConn, err := p.Acquire()
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}
	defer poolConn.Release()

	conn := poolConn.Conn()

	page, continuation, persist, cursorID, err := documentdb_api.FindCursorFirstPage(ctx, conn, p.l, db, spec, 0)
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}

	p.l.DebugContext(
		ctx, "Find result",
		slog.Any("page", page), slog.Any("continuation", continuation),
		slog.Bool("persist", persist), slog.Int64("cursor", cursorID),
	)

	if persist {
		conn = poolConn.hijack()
	} else {
		conn = nil
	}

	p.r.NewCursor(cursorID, continuation, conn)

	return page, cursorID, nil
}

// Aggregate returns the first page of the `aggregate` cursor and the cursor ID.
func (p *Pool) Aggregate(ctx context.Context, db string, spec wirebson.RawDocument) (wirebson.RawDocument, int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "pool.Aggregate")
	defer span.End()

	poolConn, err := p.Acquire()
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}
	defer poolConn.Release()

	conn := poolConn.Conn()

	page, continuation, persist, cursorID, err := documentdb_api.AggregateCursorFirstPage(ctx, conn, p.l, db, spec, 0)
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}

	p.l.DebugContext(
		ctx, "Aggregate result",
		slog.Any("page", page), slog.Any("continuation", continuation),
		slog.Bool("persist", persist), slog.Int64("cursor", cursorID),
	)

	if persist {
		conn = poolConn.hijack()
	} else {
		conn = nil
	}

	p.r.NewCursor(cursorID, continuation, conn)

	return page, cursorID, nil
}

// ListIndexes returns the first page of the `listIndexes` cursor and the cursor ID.
func (p *Pool) ListIndexes(ctx context.Context, db string, spec wirebson.RawDocument) (wirebson.RawDocument, int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "pool.ListIndexes")
	defer span.End()

	poolConn, err := p.Acquire()
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}

	defer poolConn.Release()

	conn := poolConn.Conn()

	page, continuation, persist, cursorID, err := documentdb_api.ListIndexesCursorFirstPage(ctx, conn, p.l, db, spec, 0)
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}

	p.l.DebugContext(
		ctx, "ListIndexes result",
		slog.Any("page", page), slog.Any("continuation", continuation),
		slog.Bool("persist", persist), slog.Int64("cursor", cursorID),
	)

	if persist {
		conn = poolConn.hijack()
	} else {
		conn = nil
	}

	p.r.NewCursor(cursorID, continuation, conn)

	return page, cursorID, nil
}
