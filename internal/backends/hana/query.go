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

package hana

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func prepareSelectClause(schema, table string) string {
	return fmt.Sprintf("SELECT * FROM %q.%q", schema, table)
}

func makeFilter(key, op string, value any) string {
	var valStr string

	switch v := value.(type) {
	case *types.Document, *types.Array, types.Binary,
		types.NullType, types.Regex, types.Timestamp:
	// type not supported for pushdown
	case bool:
		valStr = fmt.Sprintf("TO_JSON_BOOLEAN(%t)", value)
	case int32, int64:
		valStr = fmt.Sprintf("%d", value)
	case float64:
		// TODO check for MaxSafeValues
		valStr = fmt.Sprintf("%f", value)
	case nil:
		valStr = "NULL"
	case string, types.ObjectID, time.Time:
		marshaledValue := string(must.NotFail(sjson.MarshalSingleValue(v)))
		valStr = strings.ReplaceAll(marshaledValue, "\"", "'")
	default:
		panic(fmt.Sprintf("Unexpected type of value: %v", v))
	}

	res := fmt.Sprintf("%q %s %s", key, op, valStr)

	return res
}

func prepareWhereClause(filter *types.Document) (string, error) {
	var filters []string

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

		// don't pushdown $comment, as it's attached to query with select clause
		//
		// all of the other top-level operators such as `$or` do not support pushdown yet
		if strings.HasPrefix(rootKey, "$") {
			continue
		}

		path, err := types.NewPathFromString(rootKey)

		var pe *types.PathError

		switch {
		case err == nil:
			// Handle dot notation.
			// TODO https://github.com/FerretDB/FerretDB/issues/2069
			if path.Len() > 1 {
				continue
			}
		case errors.As(err, &pe):
			// ignore empty key error, otherwise return error
			if pe.Code() != types.ErrPathElementEmpty {
				return "", lazyerrors.Error(err)
			}
		default:
			panic("Invalid error type: PathError expected")
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
					if f := makeFilter(rootKey, "=", v); f != "" {
						filters = append(filters, f)
					}

				case "$ne":
					if f := makeFilter(rootKey, "<>", v); f != "" {
						filters = append(filters, f)
					}
				default:
					// $gt and $lt
					// TODO https://github.com/FerretDB/FerretDB/issues/1875
					continue
				}
			}

		case *types.Array, types.Binary, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown

		case float64, string, types.ObjectID, bool, time.Time, int32, int64:
			if f := makeFilter(rootKey, "=", v); f != "" {
				filters = append(filters, f)
			}

		default:
			panic(fmt.Sprintf("Unexpected type of value: %v", v))
		}
	}

	whereClause := ""
	if len(filters) > 0 {
		whereClause = " WHERE " + strings.Join(filters, " AND ")
	}

	return whereClause, nil
}

func prepareOrderByClause(sort *backends.SortField) (string, error) {
	if sort == nil {
		return "", nil
	}

	var order string
	order = "ASC"
	if sort.Descending {
		order = "DESC"
	}

	orderByClause := fmt.Sprintf(" ORDER BY %q %s", sort.Key, order)

	return orderByClause, nil
}
