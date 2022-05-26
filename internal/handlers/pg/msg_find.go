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
	"fmt"
	"math"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
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
	common.Ignored(document, h.l, ignoredFields...)

	var filter, sort, projection *types.Document
	if filter, err = common.GetOptionalParam(document, "filter", filter); err != nil {
		return nil, err
	}
	if sort, err = common.GetOptionalParam(document, "sort", sort); err != nil {
		return nil, common.NewErrorMsg(common.ErrTypeMismatch, "Expected field sort to be of type object")
	}
	if projection, err = common.GetOptionalParam(document, "projection", projection); err != nil {
		return nil, err
	}

	maxTimeMS, err := getMaxTimeMSParameter(document)
	if err != nil {
		return nil, err
	}

	runCtx, runCancel := context.WithTimeout(ctx, time.Duration(maxTimeMS)*time.Millisecond)
	defer runCancel()
	if maxTimeMS == 0 {
		runCtx = ctx
	}

	var limit int64
	if l, _ := document.Get("limit"); l != nil {
		if limit, err = common.GetWholeNumberParam(l); err != nil {
			return nil, err
		}
	}

	var sp sqlParam
	if sp.db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}
	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}
	var ok bool
	if sp.collection, ok = collectionParam.(string); !ok {
		return nil, common.NewErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
		)
	}

	// get comment from options.FindOne().SetComment() method
	if sp.comment, err = common.GetOptionalParam(document, "comment", sp.comment); err != nil {
		return nil, err
	}
	// get comment from query, e.g. db.collection.find({$comment: "test"})
	if filter != nil {
		if sp.comment, err = common.GetOptionalParam(filter, "$comment", sp.comment); err != nil {
			return nil, err
		}
	}

	fetchedDocs, err := h.fetch(runCtx, sp)
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
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"firstBatch", firstBatch,
				"id", int64(0), // TODO
				"ns", sp.db+"."+sp.collection,
			)),
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// getMaxTimeMSParameter gets maxTimeMS optional parameter.
func getMaxTimeMSParameter(document *types.Document) (int32, error) {
	var maxTimeMS int32
	maxTimeMSParam, err := common.GetOptionalParam(document, "maxTimeMS", maxTimeMS)
	if err != nil {
		return 0, getErrorForInvalidMaxTimeMS(document)
	}

	if maxTimeMSParam < 0 {
		return 0, common.NewErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("%v value for maxTimeMS is out of range", maxTimeMSParam),
		)
	}

	return maxTimeMSParam, nil
}

// getErrorForInvalidMaxTimeMS returns error for invalid maxTimeMS type.
func getErrorForInvalidMaxTimeMS(document *types.Document) error {
	v, err := document.Get("maxTimeMS")
	if err != nil {
		return nil
	}

	switch maxTimeMS := v.(type) {
	case float64:
		if math.IsInf(maxTimeMS, -1) {
			return common.NewErrorMsg(
				common.ErrBadValue,
				fmt.Sprintf("%v value for maxTimeMS is out of range", math.MinInt64),
			)
		}

		if math.IsInf(maxTimeMS, +1) {
			return common.NewErrorMsg(
				common.ErrBadValue,
				fmt.Sprintf("%v value for maxTimeMS is out of range", math.MaxInt64),
			)
		}

		if maxTimeMS > math.MaxInt32 || maxTimeMS < math.MinInt32 {
			return common.NewErrorMsg(
				common.ErrBadValue,
				fmt.Sprintf("%v value for maxTimeMS is out of range", int64(maxTimeMS)),
			)
		}

		if maxTimeMS == math.Trunc(maxTimeMS) {
			return nil
		}

		return common.NewErrorMsg(common.ErrBadValue, "maxTimeMS must be an integer")
	case int64:
		if maxTimeMS < math.MaxInt32 && maxTimeMS > math.MinInt32 {
			return nil
		}

		return common.NewErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("%v value for maxTimeMS is out of range", v),
		)
	default:
		return common.NewErrorMsg(common.ErrBadValue, "maxTimeMS must be a number")
	}
}
