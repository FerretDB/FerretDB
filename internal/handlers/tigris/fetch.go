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

	"github.com/tigrisdata/tigrisdb-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// fetch fetches all documents from the given database and collection.
//
// TODO https://github.com/FerretDB/FerretDB/issues/372
func (h *Handler) fetch(ctx context.Context, db, collection string) ([]*types.Document, error) {
	iterator, err := h.client.conn.Read(ctx, db, collection, driver.Filter("{}"))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res []*types.Document
	elem := new(driver.Document)
	for iterator.Next(elem) {
		var anyDoc map[string]any
		err := json.Unmarshal(*elem, &anyDoc)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		doc, err := types.NewDocumentFromMap(anyDoc)
		if err != nil {
			return nil, err
		}
		res = append(res, doc)
	}
	return res, nil
}
