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
	"encoding/json"
	"fmt"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDelete implements HandlerInterface.
func (h *Handler) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "comment") // TODO https://github.com/FerretDB/FerretDB/issues/849
	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}
	common.Ignored(document, h.L, "ordered") // TODO https://github.com/FerretDB/FerretDB/issues/848
	common.Ignored(document, h.L, "writeConcern")

	var deletes *types.Array
	if deletes, err = common.GetOptionalParam(document, "deletes", deletes); err != nil {
		return nil, err
	}

	var deleted int32
	for i := 0; i < deletes.Len(); i++ {
		d, err := common.AssertType[*types.Document](must.NotFail(deletes.Get(i)))
		if err != nil {
			return nil, err
		}

		if err := common.Unimplemented(d, "collation", "hint"); err != nil {
			return nil, err
		}

		var filter *types.Document
		if filter, err = common.GetOptionalParam(d, "q", filter); err != nil {
			return nil, err
		}

		var limit int64

		l, err := d.Get("limit")
		if err != nil {
			return nil, common.NewErrorMsg(
				common.ErrMissingField,
				"BSON field 'delete.deletes.limit' is missing but a required field",
			)
		}

		if limit, err = common.GetWholeNumberParam(l); err != nil {
			return nil, err
		}

		var fp fetchParam
		if fp.db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
			return nil, err
		}
		collectionParam, err := document.Get(document.Command())
		if err != nil {
			return nil, err
		}
		var ok bool
		if fp.collection, ok = collectionParam.(string); !ok {
			return nil, common.NewErrorMsg(
				common.ErrBadValue,
				fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			)
		}

		fetchedDocs, err := h.fetch(ctx, fp)
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

		if resDocs, err = common.LimitDocuments(resDocs, limit); err != nil {
			return nil, err
		}

		if len(resDocs) == 0 {
			continue
		}

		res, err := h.delete(ctx, fp, resDocs)
		if err != nil {
			return nil, err
		}

		deleted += int32(res)
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"n", deleted,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// delete deletes documents by _id.
func (h *Handler) delete(ctx context.Context, fp fetchParam, docs []*types.Document) (int, error) {
	ids := make([]map[string]any, len(docs))
	for i, doc := range docs {
		id := must.NotFail(tjson.Marshal(must.NotFail(doc.Get("_id"))))
		ids[i] = map[string]any{"_id": map[string]json.RawMessage{"$eq": id}}
	}

	var f driver.Filter
	switch len(ids) {
	case 0:
		f = driver.Filter(`{}`)
	case 1:
		f = must.NotFail(json.Marshal(ids[0]))
	default:
		f = must.NotFail(json.Marshal(map[string]any{"$or": ids}))
	}

	h.L.Sugar().Debugf("Delete filter: %s", f)

	_, err := h.db.Driver.UseDatabase(fp.db).Delete(ctx, fp.collection, f)
	if err != nil {
		return 0, lazyerrors.Error(err)
	}

	return len(ids), nil
}
