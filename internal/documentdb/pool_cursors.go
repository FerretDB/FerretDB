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

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"go.opentelemetry.io/otel"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/devbuild"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// GetMore returns the next page of the cursor.
// It is a part of the implementation of the `getMore` command.
func (p *Pool) GetMore(ctx context.Context, db string, spec wirebson.RawDocument, cursorID int64) (wirebson.RawDocument, error) {
	ctx, span := otel.Tracer("").Start(ctx, "documentdb.Pool.GetMore")
	defer span.End()

	continuation, conn := p.r.GetCursor(cursorID)
	if continuation == nil {
		if p.r.ConsumeInvalidation(cursorID) {
			// QueryPlanKilled is not yet in the upstream generated code list.
			err := mongoerrors.New(mongoerrors.Code(175), "Collection was renamed")
			err.Name = "QueryPlanKilled"
			return nil, err
		}
		return nil, mongoerrors.New(
			mongoerrors.ErrCursorNotFound,
			fmt.Sprintf("cursor id %d not found", cursorID),
		)
	}

	if conn == nil {
		poolConn, err := p.AcquireContext(ctx)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		defer poolConn.Release()

		conn = poolConn.Conn()
	}

	if err := p.checkCursorGeneration(ctx, conn, cursorID); err != nil {
		return nil, err
	}
	page, continuation, err := documentdb_api.CursorGetMore(ctx, conn, p.l, db, spec, continuation)
	if err != nil {
		p.r.CloseCursor(ctx, cursorID)
		return nil, lazyerrors.Error(err)
	}
	// A rename may commit while the page is being fetched. Check again before
	// exposing it, including when the cursor has just reached exhaustion.
	if err := p.checkCursorGeneration(ctx, conn, cursorID); err != nil {
		return nil, err
	}

	p.l.DebugContext(
		ctx, "GetMore result", slog.Int64("id", cursorID),
		slog.Any("page", logging.LazyDecoder(page)), slog.Any("continuation", logging.LazyDeepDecoder(continuation)),
	)

	p.r.UpdateCursor(ctx, cursorID, continuation)

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
	ctx, span := otel.Tracer("").Start(ctx, "documentdb.Pool.KillCursor")
	defer span.End()

	return p.r.CloseCursor(ctx, id)
}

// ListCollections returns the first page of the `listCollections` cursor and the cursor ID.
func (p *Pool) ListCollections(ctx context.Context, db string, spec wirebson.RawDocument) (wirebson.RawDocument, int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "documentdb.Pool.ListCollections")
	defer span.End()

	poolConn, err := p.AcquireContext(ctx)
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
		slog.Any("page", logging.LazyDecoder(page)), slog.Any("continuation", logging.LazyDeepDecoder(continuation)),
		slog.Bool("persist", persist), slog.Int64("id", cursorID),
	)

	if p.shouldStore(ctx, page, continuation, cursorID) {
		var release func()
		if persist {
			conn, release, err = p.persistCursor(ctx, poolConn)
			if err != nil {
				return nil, 0, err
			}
		} else {
			conn = nil
		}

		p.r.NewCursor(ctx, cursorID, continuation, conn, db, "", release)
	}

	return page, cursorID, nil
}

// Find returns the first page of the `find` cursor and the cursor ID.
func (p *Pool) Find(ctx context.Context, db string, spec wirebson.RawDocument) (wirebson.RawDocument, int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "documentdb.Pool.Find")
	defer span.End()

	poolConn, err := p.AcquireContext(ctx)
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}
	defer poolConn.Release()

	conn := poolConn.Conn()
	generation, err := p.cursorGeneration(ctx, conn, db, cursorCollection(spec))
	if err != nil {
		return nil, 0, err
	}

	page, continuation, persist, cursorID, err := documentdb_api.FindCursorFirstPage(ctx, conn, p.l, db, spec, 0)
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}

	p.l.DebugContext(
		ctx, "Find result",
		slog.Any("page", logging.LazyDecoder(page)), slog.Any("continuation", logging.LazyDeepDecoder(continuation)),
		slog.Bool("persist", persist), slog.Int64("id", cursorID),
	)

	if p.shouldStore(ctx, page, continuation, cursorID) {
		var release func()
		if persist {
			conn, release, err = p.persistCursor(ctx, poolConn)
			if err != nil {
				return nil, 0, err
			}
		} else {
			conn = nil
		}

		p.r.NewCursor(ctx, cursorID, continuation, conn, db, cursorCollection(spec), release)
		if p.unifiCompat && cursorCollection(spec) != "" {
			p.r.SetGeneration(cursorID, generation)
		}
	}

	return page, cursorID, nil
}

// Aggregate returns the first page of the `aggregate` cursor and the cursor ID.
func (p *Pool) Aggregate(ctx context.Context, db string, spec wirebson.RawDocument) (wirebson.RawDocument, int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "documentdb.Pool.Aggregate")
	defer span.End()

	poolConn, err := p.AcquireContext(ctx)
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}
	defer poolConn.Release()

	conn := poolConn.Conn()
	generation, err := p.cursorGeneration(ctx, conn, db, cursorCollection(spec))
	if err != nil {
		return nil, 0, err
	}

	dependencies, err := p.aggregateDependencies(ctx, conn, db, spec)
	if err != nil {
		return nil, 0, err
	}

	page, continuation, persist, cursorID, err := documentdb_api.AggregateCursorFirstPage(ctx, conn, p.l, db, spec, 0)
	if err != nil {
		return nil, 0, lazyerrors.Error(err)
	}

	p.l.DebugContext(
		ctx, "Aggregate result",
		slog.Any("page", logging.LazyDecoder(page)), slog.Any("continuation", logging.LazyDeepDecoder(continuation)),
		slog.Bool("persist", persist), slog.Int64("id", cursorID),
	)

	if p.shouldStore(ctx, page, continuation, cursorID) {
		var release func()
		if persist {
			conn, release, err = p.persistCursor(ctx, poolConn)
			if err != nil {
				return nil, 0, err
			}
		} else {
			conn = nil
		}

		p.r.NewCursor(ctx, cursorID, continuation, conn, db, cursorCollection(spec), release)
		p.r.SetDependencies(cursorID, dependencies)
		if p.unifiCompat && cursorCollection(spec) != "" {
			p.r.SetGeneration(cursorID, generation)
		}
	}

	return page, cursorID, nil
}

// ListIndexes returns the first page of the `listIndexes` cursor and the cursor ID.
func (p *Pool) ListIndexes(ctx context.Context, db string, spec wirebson.RawDocument) (wirebson.RawDocument, int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "documentdb.Pool.ListIndexes")
	defer span.End()

	poolConn, err := p.AcquireContext(ctx)
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
		slog.Any("page", logging.LazyDecoder(page)), slog.Any("continuation", logging.LazyDeepDecoder(continuation)),
		slog.Bool("persist", persist), slog.Int64("id", cursorID),
	)

	if p.shouldStore(ctx, page, continuation, cursorID) {
		var release func()
		if persist {
			conn, release, err = p.persistCursor(ctx, poolConn)
			if err != nil {
				return nil, 0, err
			}
		} else {
			conn = nil
		}

		p.r.NewCursor(ctx, cursorID, continuation, conn, db, cursorCollection(spec), release)
	}

	return page, cursorID, nil
}

func cursorCollection(spec wirebson.RawDocument) string {
	doc, err := spec.Decode()
	if err != nil {
		return ""
	}
	name, _ := doc.Get(doc.Command()).(string)
	return name
}

// CloseNamespaceCursors invalidates source/target cursors before a rename.
func (p *Pool) CloseNamespaceCursors(ctx context.Context, database, collection string) {
	p.r.CloseNamespace(ctx, database, collection)
}

// shouldStore returns true if cursor should be stored.
//
// It logs when persist is true but cursorID is zero or continuation is empty.
// In debug builds, it also checks that cursorID matches the one in the page.
// Both checks are workarounds for
// TODO https://github.com/FerretDB/FerretDB/issues/5445
func (p *Pool) shouldStore(ctx context.Context, page, cont wirebson.RawDocument, cursorID int64) bool {
	if devbuild.Enabled {
		pageDoc := must.NotFail(page.Decode())

		cursor, ok := pageDoc.Get("cursor").(wirebson.AnyDocument)
		if !ok {
			if cursorID != 0 {
				p.l.LogAttrs(
					ctx, logging.LevelDPanic, "cursorID is not zero but cursor is missing",
					slog.Int64("id", cursorID), slog.Any("page", pageDoc),
				)
			}
		}

		cursorDoc := must.NotFail(cursor.Decode())
		p.l.DebugContext(
			ctx, "Decoded cursor",
			slog.Int64("id", cursorID), slog.Any("cursor", cursorDoc),
		)

		id, ok := cursorDoc.Get("id").(int64)
		if !ok {
			if cursorID != 0 {
				p.l.LogAttrs(
					ctx, logging.LevelDPanic, "cursorID is not zero but cursor.id is missing",
					slog.Int64("id", cursorID), slog.Any("cursor", cursorDoc),
				)
			}
		}

		if id != cursorID {
			p.l.LogAttrs(
				ctx, logging.LevelDPanic, "cursorID does not match cursor.id",
				slog.Int64("id", cursorID), slog.Any("cursor", cursorDoc),
			)
		}
	}

	if cursorID == 0 {
		p.l.DebugContext(ctx, "Not storing cursor with zero ID")
		return false
	}

	if len(cont) == 0 {
		p.l.DebugContext(ctx, "Not storing cursor with empty continuation")
		return false
	}

	return true
}
