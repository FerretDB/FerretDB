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

	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// FetchParam represents options/parameters used by the fetch/query.
type FetchParam struct {
	DB         string
	Collection string
}

func (tdb *TigrisDB) QueryDocuments(ctx context.Context, param FetchParam) ([]*types.Document, error) {
	db := tdb.Driver.UseDatabase(param.DB)

	collection, err := db.DescribeCollection(ctx, param.Collection)
	switch err := err.(type) {
	case nil:
		// do nothing
	case *driver.Error:
		if IsNotFound(err) {
			tdb.L.Debug(
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

	iter, err := db.Read(ctx, param.Collection, driver.Filter(`{}`), nil)
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
