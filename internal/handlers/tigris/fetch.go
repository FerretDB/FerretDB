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

	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// fetchParam represents options/parameters used by the fetch.
type fetchParam struct {
	db         string
	collection string
}

// fetch fetches all documents from the given database and collection.
//
// TODO https://github.com/FerretDB/FerretDB/issues/372
func (h *Handler) fetch(ctx context.Context, param fetchParam) ([]*types.Document, error) {
	db := h.db.Driver.UseDatabase(param.db)

	collection, err := db.DescribeCollection(ctx, param.collection)
	switch err := err.(type) {
	case nil:
		// do nothing
	case *driver.Error:
		if tigrisdb.IsNotFound(err) {
			h.L.Debug(
				"Collection doesn't exist, handling a case to deal with a non-existing collection (return empty list)",
				zap.String("db", param.db), zap.String("collection", param.collection),
			)
			return []*types.Document{}, nil
		}
		return nil, lazyerrors.Error(err)
	default:
		return nil, lazyerrors.Error(err)
	}

	var schema tjson.Schema
	if err = schema.Unmarshal(collection.Schema); err != nil {
		return nil, lazyerrors.Error(err)
	}

	iter, err := db.Read(ctx, param.collection, driver.Filter(`{}`), nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer iter.Close()

	var res []*types.Document
	var d driver.Document
	for iter.Next(&d) {
		doc, err := tjson.Unmarshal(d, &schema)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res = append(res, doc.(*types.Document))
	}

	return res, iter.Err()
}
