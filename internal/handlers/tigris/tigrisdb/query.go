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

package tigrisdb

import (
	"context"
	"encoding/json"

	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// FetchParam represents options/parameters used by the fetch/query.
type FetchParam struct {
	DB         string
	Collection string

	// Query filter for possible pushdown; may be ignored in part or entirely.
	Filter *types.Document
}

// QueryDocuments fetches documents from the given collection.
func (tdb *TigrisDB) QueryDocuments(ctx context.Context, param *FetchParam) (iterator.Interface[int, *types.Document], error) {
	db := tdb.Driver.UseDatabase(param.DB)

	collection, err := db.DescribeCollection(ctx, param.Collection)
	switch err := err.(type) {
	case nil:
		// do nothing
	case *driver.Error:
		if IsNotFound(err) {
			tdb.l.Debug(
				"Collection doesn't exist, handling a case to deal with a non-existing collection (return empty list)",
				zap.String("db", param.DB), zap.String("collection", param.Collection),
			)

			return nil, nil
		}

		return nil, lazyerrors.Error(err)
	default:
		return nil, lazyerrors.Error(err)
	}

	var schema tjson.Schema
	if err = schema.Unmarshal(collection.Schema); err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter := tdb.BuildFilter(param.Filter)
	tdb.l.Sugar().Debugf("Read filter: %s", filter)

	tigrisIter, err := db.Read(ctx, param.Collection, filter, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	iter := newQueryIterator(tigrisIter, &schema)

	return iter, nil
}

// BuildFilter returns Tigris filter expression that may cover a part of the given filter.
//
// FerretDB always filters data itself, so that should be a purely performance optimization.
func (tdb *TigrisDB) BuildFilter(filter *types.Document) driver.Filter {
	res := map[string]any{}

	for k, v := range filter.Map() {
		// filter only by _id for now
		if k != "_id" {
			continue
		}

		switch v.(type) {
		case string:
			// filtering by string values is complicated if the storage supports encodings, collations, etc,
			// but Tigris does not support any of these
		case types.ObjectID:
			// filtering by ObjectID is always safe
		default:
			// skip other types for now
			continue
		}

		// filter by the exact _id value
		id := must.NotFail(tjson.Marshal(v))
		res["_id"] = json.RawMessage(id)
	}

	return must.NotFail(json.Marshal(res))
}
