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
	"time"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
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

	params, err := common.GetFindParams(document, h.L)
	if err != nil {
		return nil, err
	}

	if params.MaxTimeMS != 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
		defer cancel()

		ctx = ctxWithTimeout
	}

	fp := tigrisdb.FetchParam{
		DB:         params.DB,
		Collection: params.Collection,
		Filter:     params.Filter,
	}

	resDocs, err := h.fetchAndFilterDocs(ctx, &fp)
	if err != nil {
		return nil, err
	}

	if err = common.SortDocuments(resDocs, params.Sort); err != nil {
		return nil, err
	}

	if resDocs, err = common.LimitDocuments(resDocs, params.Limit); err != nil {
		return nil, err
	}

	if err = common.ProjectDocuments(resDocs, params.Projection); err != nil {
		return nil, err
	}

	batchSize := len(resDocs)
	if len(resDocs) > int(params.BatchSize) {
		batchSize = int(params.BatchSize)
	}

	firstBatch := types.MakeArray(batchSize)
	resultDocumentsArray := types.MakeArray(0)

	for i := 0; i < len(resDocs); i++ {
		if i < batchSize {
			firstBatch.Append(resDocs[i])
		} else {
			resultDocumentsArray.Append(resDocs[i])
		}
	}

	if resultDocumentsArray.Len() > 0 {
		conninfo.Get(ctx).SetCursor(fp.DB+"."+fp.Collection, resultDocumentsArray.Iterator())
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"id", int64(1),
				"ns", fp.DB+"."+fp.Collection,
				"firstBatch", firstBatch,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// fetchAndFilterDocs fetches documents from the database and filters them using the provided FetchParam.Filter.
func (h *Handler) fetchAndFilterDocs(ctx context.Context, fp *tigrisdb.FetchParam) ([]*types.Document, error) {
	fetchedDocs, err := h.db.QueryDocuments(ctx, fp)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resDocs := make([]*types.Document, 0, 16)

	for _, doc := range fetchedDocs {
		var matches bool

		if matches, err = common.FilterDocument(doc, fp.Filter); err != nil {
			return nil, err
		}

		if !matches {
			continue
		}

		resDocs = append(resDocs, doc)
	}

	return resDocs, nil
}
