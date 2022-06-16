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

	"github.com/tigrisdata/tigris-client-go/fields"
	"github.com/tigrisdata/tigris-client-go/filter"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUpdate implements HandlerInterface.
func (h *Handler) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}
	common.Ignored(document, h.L, "ordered", "writeConcern", "bypassDocumentValidation", "comment")

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

	var updates *types.Array
	if updates, err = common.GetOptionalParam(document, "updates", updates); err != nil {
		return nil, err
	}

	// created, err := h.pgPool.CreateTableIfNotExist(ctx, fp.db, fp.collection)
	// if err != nil {
	// 	return nil, err
	// }
	// if created {
	// 	h.L.Info("Created table.", zap.String("schema", fp.db), zap.String("table", fp.collection))
	// }

	var matched, modified int32
	var upserted types.Array
	for i := 0; i < updates.Len(); i++ {
		update, err := common.AssertType[*types.Document](must.NotFail(updates.Get(i)))
		if err != nil {
			return nil, err
		}

		unimplementedFields := []string{
			"c",
			"multi",
			"collation",
			"arrayFilters",
			"hint",
		}
		if err := common.Unimplemented(update, unimplementedFields...); err != nil {
			return nil, err
		}

		var q, u *types.Document
		var upsert bool
		if q, err = common.GetOptionalParam(update, "q", q); err != nil {
			return nil, err
		}
		if u, err = common.GetOptionalParam(update, "u", u); err != nil {
			return nil, err
		}
		if u != nil {
			if err = common.ValidateUpdateOperators(u); err != nil {
				return nil, err
			}
		}

		if upsert, err = common.GetOptionalParam(update, "upsert", upsert); err != nil {
			return nil, err
		}

		fetchedDocs, err := h.fetch(ctx, fp)
		if err != nil {
			return nil, err
		}

		resDocs := make([]*types.Document, 0, 16)
		for _, doc := range fetchedDocs {
			matches, err := common.FilterDocument(doc, q)
			if err != nil {
				return nil, err
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)
		}

		if len(resDocs) == 0 {
			if !upsert {
				// nothing to do, continue to the next update operation
				continue
			}

			doc := q.DeepCopy()
			if _, err = common.UpdateDocument(doc, u); err != nil {
				return nil, err
			}
			if !doc.Has("_id") {
				must.NoError(doc.Set("_id", types.NewObjectID()))
			}

			must.NoError(upserted.Append(must.NotFail(types.NewDocument(
				"index", int32(0), // TODO
				"_id", must.NotFail(doc.Get("_id")),
			))))

			if err = h.insert(ctx, fp, doc); err != nil {
				return nil, err
			}

			matched++
			continue
		}

		matched += int32(len(resDocs))

		for _, doc := range resDocs {
			changed, err := common.UpdateDocument(doc, u)
			if err != nil {
				return nil, err
			}

			if !changed {
				continue
			}

			res, err := h.update(ctx, fp, doc)
			if err != nil {
				return nil, err
			}
			modified += int32(res)
		}
	}

	res := must.NotFail(types.NewDocument(
		"n", matched,
	))
	if upserted.Len() != 0 {
		must.NoError(res.Set("upserted", &upserted))
	}
	must.NoError(res.Set("nModified", modified))
	must.NoError(res.Set("ok", float64(1)))

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{res},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// update updates documents by _id.
func (h *Handler) update(ctx context.Context, sp fetchParam, doc *types.Document) (int, error) {
	id := must.NotFail(doc.Get("_id")).(types.ObjectID)
	f := must.NotFail(filter.Eq("_id", tjson.ObjectID(id)).Build())
	h.L.Sugar().Debugf("Filter: %s", f)

	update := fields.UpdateBuilder()
	for _, k := range doc.Keys() {
		v := must.NotFail(doc.Get(k))
		update.Set(k, json.RawMessage(must.NotFail(tjson.Marshal(v))))
	}
	u := must.NotFail(update.Build()).Built()
	h.L.Sugar().Debugf("Update: %s", u)

	res, err := h.driver.UseDatabase(sp.db).Update(ctx, sp.collection, f, u)
	if err != nil {
		return 0, lazyerrors.Error(err)
	}

	return int(res.ModifiedCount), nil
}
