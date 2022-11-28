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

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFind implements HandlerInterface.
func (h *Handler) MsgFind(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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
		"readConcern",
		"max",
		"min",
	}
	common.Ignored(document, h.L, ignoredFields...)

	var filter, sort, projection *types.Document
	if filter, err = common.GetOptionalParam(document, "filter", filter); err != nil {
		return nil, err
	}
	if sort, err = common.GetOptionalParam(document, "sort", sort); err != nil {
		return nil, common.NewCommandErrorMsgWithArgument(
			common.ErrTypeMismatch,
			"Expected field sort to be of type object",
			"sort",
		)
	}
	if projection, err = common.GetOptionalParam(document, "projection", projection); err != nil {
		return nil, err
	}

	maxTimeMS, err := common.GetOptionalPositiveNumber(document, "maxTimeMS")
	if err != nil {
		return nil, err
	}

	if maxTimeMS != 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(maxTimeMS)*time.Millisecond)
		defer cancel()

		ctx = ctxWithTimeout
	}

	var limit int64
	if l, _ := document.Get("limit"); l != nil {
		if limit, err = common.GetWholeNumberParam(l); err != nil {
			return nil, err
		}
	}

	sp := pgdb.SQLParam{
		Filter: filter,
	}

	if sp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if sp.Collection, ok = collectionParam.(string); !ok {
		return nil, common.NewCommandErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
		)
	}

	// get comment from options.FindOne().SetComment() method
	if sp.Comment, err = common.GetOptionalParam(document, "comment", sp.Comment); err != nil {
		return nil, err
	}
	// get comment from query, e.g. db.collection.find({$comment: "test"})
	if filter != nil {
		if sp.Comment, err = common.GetOptionalParam(filter, "$comment", sp.Comment); err != nil {
			return nil, err
		}
	}

	resDocs := make([]*types.Document, 0, 16)
	err = h.PgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		it, ierr := h.PgPool.GetDocuments(ctx, tx, &sp)
		if ierr != nil {
			return ierr
		}

		if it == nil {
			// no documents found
			return nil
		}

		defer it.Close()

		for {
			var doc *types.Document
			_, doc, err = it.Next()

			switch {
			case err == nil:
				// do nothing
			case errors.Is(ierr, iterator.ErrIteratorDone):
				// no more documents
				return nil
			case errors.Is(ierr, context.Canceled), errors.Is(ierr, context.DeadlineExceeded):
				h.L.Warn(
					"context canceled, stopping fetching",
					zap.String("db", sp.DB), zap.String("collection", sp.Collection), zap.Error(ierr),
				)
				return nil
			default:
				return ierr
			}

			var matches bool
			matches, err = common.FilterDocument(doc, filter)
			if err != nil {
				return err
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)
		}
	})

	if err != nil {
		return nil, err
	}

	if err = common.SortDocuments(resDocs, sort); err != nil {
		return nil, err
	}
	if resDocs, err = common.LimitDocuments(resDocs, limit); err != nil {
		return nil, err
	}
	if err = common.ProjectDocuments(resDocs, projection); err != nil {
		return nil, err
	}

	firstBatch := types.MakeArray(len(resDocs))
	for _, doc := range resDocs {
		if err = firstBatch.Append(doc); err != nil {
			return nil, err
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"firstBatch", firstBatch,
				"id", int64(0), // TODO
				"ns", sp.DB+"."+sp.Collection,
			)),
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
