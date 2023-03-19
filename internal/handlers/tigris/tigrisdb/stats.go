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

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// CollectionStats describes statistics for a Tigris collection.
type CollectionStats struct {
	NumObjects int32
	Size       int64
}

// FetchStats returns a set of statistics for the given database and collection.
func FetchStats(ctx context.Context, db driver.Database, collection string) (*CollectionStats, error) {
	collection = EncodeCollName(collection)

	info, err := db.DescribeCollection(ctx, collection)
	switch err := err.(type) {
	case nil:
		// do nothing

	case *driver.Error:
		if IsNotFound(err) {
			// If DB or collection doesn't exist just return empty stats.
			stats := &CollectionStats{
				NumObjects: 0,
				Size:       0,
			}

			return stats, nil
		}

		return nil, lazyerrors.Error(err)

	default:
		return nil, lazyerrors.Error(err)
	}

	iter, err := db.Read(ctx, collection, driver.Filter(`{}`), nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer iter.Close()

	var count int32
	var d driver.Document

	for iter.Next(&d) {
		count++
	}

	return &CollectionStats{NumObjects: count, Size: info.Size}, iter.Err()
}
