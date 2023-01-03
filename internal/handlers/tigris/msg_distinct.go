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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDistinct implements HandlerInterface.
func (h *Handler) MsgDistinct(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	unimplementedFields := []string{
		"collation",
	}
	if err = common.Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
	}

	ignoredFields := []string{
		"readConcern",
		"comment", // TODO: implement
	}
	common.Ignored(document, h.L, ignoredFields...)

	var fp tigrisdb.FetchParam

	if fp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var key string

	if key, err = common.GetRequiredParam[string](document, "key"); err != nil {
		return nil, err
	}

	if key == "" {
		return nil, common.NewCommandErrorMsg(common.ErrEmptyFieldPath,
			"FieldPath cannot be constructed with empty string",
		)
	}

	if fp.Filter, err = common.GetOptionalParam[*types.Document](document, "query", nil); err != nil {
		return nil, err
	}

	var ok bool
	if fp.Collection, ok = collectionParam.(string); !ok {
		return nil, common.NewCommandErrorMsgWithArgument(
			common.ErrInvalidNamespace,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	fetchedDocs, err := h.db.QueryDocuments(ctx, &fp)
	if err != nil {
		return nil, err
	}

	distinct := make([]any, 0, 16)
	duplicateChecker := make(map[any]struct{}, 16)

	for _, doc := range fetchedDocs {
		var matches bool
		matches, err = common.FilterDocument(doc, fp.Filter)

		if err != nil {
			return nil, err
		}

		if !matches {
			continue
		}

		var val any
		val, err = doc.Get(key)

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if _, ok := duplicateChecker[val]; ok {
			continue
		}

		duplicateChecker[val] = struct{}{}
		distinct = append(distinct, val)
	}

	// sort.Sort(distinct.Sorter())

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"values", distinct,
			"ok", float64(1),
		))},
	})

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
