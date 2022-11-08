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

package tigris

import (
	"context"
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/types"
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

	fp := tigrisdb.FetchParam{
		Filter: filter,
	}

	if fp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if fp.Collection, ok = collectionParam.(string); !ok {
		return nil, common.NewCommandErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
		)
	}

	fetchedDocs, err := h.db.QueryDocuments(ctx, &fp)
	if err != nil {
		return nil, err
	}

	resDocs := make([]*types.Document, 0, 16)
	for _, doc := range fetchedDocs {
		matches, err := common.FilterDocument(doc, filter)
		if err != nil {
			return nil, err
		}

		if !matches {
			continue
		}

		resDocs = append(resDocs, doc)
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
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"firstBatch", firstBatch,
				"id", int64(0), // TODO
				"ns", fp.DB+"."+fp.Collection,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
