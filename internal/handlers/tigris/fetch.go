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

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// fetch fetches all documents from the given database and collection.
//
// TODO https://github.com/FerretDB/FerretDB/issues/372
func (h *Handler) fetch(ctx context.Context, db, collection string) ([]*types.Document, error) {
	iterator, err := h.client.conn.Read(ctx, db, collection, driver.Filter("{}"), driver.Fields("{}"))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res []*types.Document
	for {
		var elem driver.Document
		ok := iterator.Next(&elem)
		if !ok {
			break
		}
		anyDoc, err := fjson.Unmarshal((json.RawMessage)(elem))
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		res = append(res, anyDoc.(*types.Document))
	}

	return res, nil
}
