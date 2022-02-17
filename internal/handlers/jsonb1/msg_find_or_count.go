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

package jsonb1

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFindOrCount finds documents in a collection or view and returns a cursor to the selected documents
// or count the number of documents that matches the query filter.
func (s *storage) MsgFindOrCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	unimplementedFields := []string{
		"skip",
		"returnKey",
		"showRecordId",
		"tailable",
		"oplogReplay",
		"noCursorTimeout",
		"awaitData",
		"allowPartialResults",
		"collation",
		"allowDiskUse",
		"let",
	}
	if err := common.Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
	}
	ignoredFields := []string{
		"hint",
		"batchSize",
		"singleBatch",
		"comment",
		"maxTimeMS",
		"readConcern",
		"max",
		"min",
	}
	common.Ignored(document, s.l, ignoredFields...)

	var filter *types.Document
	var sql, collection string

	var args []any
	var placeholder pg.Placeholder

	m := document.Map()
	_, isFindOp := m["find"].(string)
	db := m["$db"].(string)

	if isFindOp {
		projectionIn, _ := m["projection"].(*types.Document)
		projectionSQL, projectionArgs, err := projection(projectionIn, &placeholder)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		args = append(args, projectionArgs...)

		collection = m["find"].(string)
		filter, _ = m["filter"].(*types.Document)
		sql = fmt.Sprintf(`SELECT %s FROM %s`, projectionSQL, pgx.Identifier{db, collection}.Sanitize())
	} else {
		collection = m["count"].(string)
		filter, _ = m["query"].(*types.Document)
		sql = fmt.Sprintf(`SELECT COUNT(*) FROM %s`, pgx.Identifier{db, collection}.Sanitize())
	}

	sort, _ := m["sort"].(*types.Document)
	limit, _ := m["limit"].(int32)

	whereSQL, whereArgs, err := where(filter, &placeholder)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	args = append(args, whereArgs...)

	sql += whereSQL

	sortMap := sort.Map()
	if len(sortMap) != 0 {
		sql += " ORDER BY"

		for i, k := range sort.Keys() {
			if i != 0 {
				sql += ","
			}

			sql += " _jsonb->" + placeholder.Next()
			args = append(args, k)

			order := sortMap[k].(int32)
			if order > 0 {
				sql += " ASC"
			} else {
				sql += " DESC"
			}
		}
	}

	switch {
	case limit == 0:
		// undefined or zero - no limit
	case limit > 0:
		sql += " LIMIT " + placeholder.Next()
		args = append(args, limit)
	default:
		// TODO https://github.com/FerretDB/FerretDB/issues/79
		return nil, common.NewErrorMsg(common.ErrNotImplemented, "find: negative limit values are not supported")
	}

	rows, err := s.pgPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	var reply wire.OpMsg
	if isFindOp { //nolint:nestif // TODO simplify
		var docs types.Array
		for {
			doc, err := nextRow(rows)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			if doc == nil {
				break
			}

			if err = docs.Append(doc); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
		err = reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{types.MustNewDocument(
				"cursor", types.MustNewDocument(
					"firstBatch", &docs,
					"id", int64(0), // TODO
					"ns", db+"."+collection,
				),
				"ok", float64(1),
			)},
		})
	} else {
		var count int32
		for rows.Next() {
			err := rows.Scan(&count)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
		// in psql, the SELECT * FROM table limit `x` ignores the value of the limit,
		// so, we need this `if` statement to support this kind of query `db.actor.find().limit(10).count()`
		if count > limit && limit != 0 {
			count = limit
		}
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		err = reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{types.MustNewDocument(
				"n", count,
				"ok", float64(1),
			)},
		})
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
