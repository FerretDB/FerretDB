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
	"fmt"
	"time"

	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
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
func (tdb *TigrisDB) QueryDocuments(ctx context.Context, param *FetchParam) ([]*types.Document, error) {
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

	filter := tdb.BuildFilter(param.Filter)
	tdb.l.Sugar().Debugf("Read filter: %s", filter)

	iter, err := db.Read(ctx, param.Collection, filter, nil)
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

	err = iter.Err()

	switch {
	case err == nil:
		fallthrough
	case IsInvalidArgument(err):
		break
	default:
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// BuildFilter returns Tigris filter expression that may cover a part of the given filter.
//
// FerretDB always filters data itself, so that should be a purely performance optimization.
func (tdb *TigrisDB) BuildFilter(filter *types.Document) driver.Filter {
	res := map[string]any{}

	for k, v := range filter.Map() {
		if k != "" {
			// don't pushdown $comment and dot notation yet
			if k[0] == '$' || types.NewPathFromString(k).Len() > 1 {
				continue
			}
		}

		// _id field supports only specific types
		if k == "_id" {
			switch v.(type) {
			case *types.Document, *types.Array, types.Binary, bool, time.Time, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown

			case float64, string, types.ObjectID, int32, int64:
				// filter by the exact _id value
				id := must.NotFail(tjson.Marshal(v))
				res["_id"] = json.RawMessage(id)
			default:
				panic(fmt.Sprintf("Unexpected type of field %s: %T", k, v))
			}

			continue
		}

		switch v.(type) {
		case *types.Document, *types.Array, types.Binary, bool, time.Time, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown
			continue
		case float64, string, types.ObjectID, int32, int64:
			rawValue := must.NotFail(tjson.Marshal(v))
			res[k] = json.RawMessage(rawValue)
		default:
			panic(fmt.Sprintf("Unexpected type of field %s: %T", k, v))
		}
	}

	return must.NotFail(json.Marshal(res))
}
