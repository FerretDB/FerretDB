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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// QueryParams represents options/parameters used by the fetch/query.
type QueryParams struct {
	Filter     *types.Document
	DB         string
	Collection string
}

// QueryDocuments fetches documents from the given collection.
func (tdb *TigrisDB) QueryDocuments(ctx context.Context, qp *QueryParams) (iterator.Interface[int, *types.Document], error) {
	db := tdb.Driver.UseDatabase(qp.DB)

	collection, err := db.DescribeCollection(ctx, qp.Collection)
	switch err := err.(type) {
	case nil:
		// do nothing
	case *driver.Error:
		if IsNotFound(err) {
			tdb.l.Debug(
				"Collection doesn't exist, handling a case to deal with a non-existing collection (return empty list)",
				zap.String("db", qp.DB), zap.String("collection", qp.Collection),
			)

			return newQueryIterator(ctx, nil, nil), nil
		}

		return nil, lazyerrors.Error(err)
	default:
		return nil, lazyerrors.Error(err)
	}

	var schema tjson.Schema
	if err = schema.Unmarshal(collection.Schema); err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter, err := BuildFilter(qp.Filter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	tdb.l.Sugar().Debugf("Read filter: %s", filter)

	tigrisIter, err := db.Read(ctx, qp.Collection, driver.Filter(filter), nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	iter := newQueryIterator(ctx, tigrisIter, &schema)

	return iter, nil
}

// BuildFilter returns Tigris filter expression (JSON object) for the given filter document.
//
// If the given filter is nil, it returns empty JSON object {}.
func BuildFilter(filter *types.Document) (string, error) {
	if filter == nil {
		return "{}", nil
	}

	res := map[string]any{}

	iter := filter.Iterator()
	defer iter.Close()

	// iterate through root document
	for {
		rootKey, rootVal, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return "", lazyerrors.Error(err)
		}

		// don't pushdown $comment, it's attached to query in handlers
		if strings.HasPrefix(rootKey, "$") {
			continue
		}

		path, err := types.NewPathFromString(rootKey)

		var pe *types.DocumentPathError

		switch {
		case err == nil:
			// TODO dot notation https://github.com/FerretDB/FerretDB/issues/2069
			if path.Len() > 1 {
				continue
			}
		case errors.As(err, &pe):
			// ignore empty key error, otherwise return error
			if pe.Code() != types.ErrDocumentPathEmptyKey {
				return "", lazyerrors.Error(err)
			}
		default:
			panic("Invalid error type: DocumentPathError expected ")
		}

		switch v := rootVal.(type) {
		case *types.Document:
			iter := v.Iterator()
			defer iter.Close()

			// iterate through subdocument, as it may contain operators
			for {
				k, v, err := iter.Next()
				if err != nil {
					if errors.Is(err, iterator.ErrIteratorDone) {
						break
					}

					return "", lazyerrors.Error(err)
				}

				switch k {
				case "$eq":
					switch docVal := v.(type) {
					case *types.Document, *types.Array, types.Binary,
						types.NullType, types.Regex, types.Timestamp:
						// type not supported for pushdown
					case float64, string, types.ObjectID, bool, time.Time, int32, int64:
						rawValue, err := tjson.Marshal(docVal)
						if err != nil {
							return "", lazyerrors.Error(err)
						}
						res[rootKey] = json.RawMessage(rawValue)
					default:
						panic(fmt.Sprintf("Unexpected type of value: %v", v))
					}

				default:
					// TODO $gt and $lt https://github.com/FerretDB/FerretDB/issues/1875
					// TODO $ne https://github.com/FerretDB/FerretDB/issues/2052
					continue
				}
			}

		case *types.Array, types.Binary, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown
			continue

		case float64, string, types.ObjectID, bool, time.Time, int32, int64:
			rawValue, err := tjson.Marshal(v)
			if err != nil {
				return "", lazyerrors.Error(err)
			}
			res[rootKey] = json.RawMessage(rawValue)

		default:
			panic(fmt.Sprintf("Unexpected type of field %s: %T", rootKey, v))
		}
	}

	result, err := json.Marshal(res)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	return string(result), nil
}
