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

package postgresql

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// prepareWhereClause adds WHERE clause with given filters to the query and returns the query and arguments.
func prepareWhereClause(p *Placeholder, sqlFilters *types.Document) (string, []any, error) {
	var filters []string
	var args []any

	iter := sqlFilters.Iterator()
	defer iter.Close()

	// iterate through root document
	for {
		rootKey, rootVal, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return "", nil, lazyerrors.Error(err)
		}

		// don't pushdown $comment, it's attached to query in handlers
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
				return "", nil, lazyerrors.Error(err)
			}
		default:
			panic("Invalid error type: PathError expected")
		}

		switch v := rootVal.(type) {
		case *types.Document:
			fs, as, err := filterOperator(p, rootKey, v)
			if err != nil {
				return "", nil, lazyerrors.Error(err)
			}

			filters = append(filters, fs...)
			args = append(args, as...)
		case *types.Array, types.Binary, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown

		case float64, string, types.ObjectID, bool, time.Time, int32, int64:
			if f, a := filterEqual(p, rootKey, v); f != "" {
				filters = append(filters, f)
				args = append(args, a...)
			}

		default:
			panic(fmt.Sprintf("Unexpected type of value: %v", v))
		}
	}

	var filter string
	if len(filters) > 0 {
		filter = `WHERE ` + strings.Join(filters, " AND ")
	}

	return filter, args, nil
}

// prepareOrderByClause adds ORDER BY clause with given sort document and returns the query and arguments.
func prepareOrderByClause(p *Placeholder, sort *types.Document) (string, []any, error) {
	iter := sort.Iterator()
	defer iter.Close()

	var key string
	var order types.SortType

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return "", nil, lazyerrors.Error(err)
		}

		// Skip sorting if there are more than one sort parameters
		if order != 0 {
			return "", nil, nil
		}

		order, err = common.GetSortType(k, v)
		if err != nil {
			return "", nil, err
		}

		key = k
	}

	// Skip sorting dot notation
	if strings.ContainsRune(key, '.') {
		return "", nil, nil
	}

	var sqlOrder string

	switch order {
	case types.Descending:
		sqlOrder = "DESC"
	case types.Ascending:
		sqlOrder = "ASC"
	case 0:
		return "", nil, nil
	default:
		panic(fmt.Sprint("forbidden order:", order))
	}

	return fmt.Sprintf(" ORDER BY _jsonb->%s %s", p.Next(), sqlOrder), []any{key}, nil
}

// filterOperator returns filters and arguments from doc.
//
// This function was extract to avoid `Possible resource leak, 'defer' is called in the 'for' loop`.
func filterOperator(p *Placeholder, rootKey string, doc *types.Document) (filters []string, args []any, err error) {
	iter := doc.Iterator()
	defer iter.Close()

	// iterate through subdocument, as it may contain operators
	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, nil, lazyerrors.Error(err)
		}

		switch k {
		case "$eq":
			if f, a := filterEqual(p, rootKey, v); f != "" {
				filters = append(filters, f)
				args = append(args, a...)
			}

		case "$ne":
			sql := `NOT ( ` +
				// does document contain the key,
				// it is necessary, as NOT won't work correctly if the key does not exist.
				`_jsonb ? %[1]s AND ` +
				// does the value under the key is equal to filter value
				`_jsonb->%[1]s @> %[2]s AND ` +
				// does the value type is equal to the filter's one
				`_jsonb->'$s'->'p'->%[1]s->'t' = '"%[3]s"' )`

			switch v := v.(type) {
			case *types.Document, *types.Array, types.Binary,
				types.NullType, types.Regex, types.Timestamp:
				// type not supported for pushdown

			case float64, bool, int32, int64:
				filters = append(filters, fmt.Sprintf(sql, p.Next(), p.Next(), sjson.GetTypeOfValue(v)))
				args = append(args, rootKey, v)

			case string, types.ObjectID, time.Time:
				filters = append(filters, fmt.Sprintf(sql, p.Next(), p.Next(), sjson.GetTypeOfValue(v)))
				args = append(args, rootKey, string(must.NotFail(sjson.MarshalSingleValue(v))))

			default:
				panic(fmt.Sprintf("Unexpected type of value: %v", v))
			}

		default:
			// $gt and $lt
			// TODO https://github.com/FerretDB/FerretDB/issues/1875
			continue
		}
	}

	return filters, args, nil
}

// filterEqual returns the proper SQL filter with arguments that filters documents
// where the value under k is equal to v.
func filterEqual(p *Placeholder, k string, v any) (filter string, args []any) {
	// Select if value under the key is equal to provided value.
	sql := `_jsonb->%[1]s @> %[2]s`

	switch v := v.(type) {
	case *types.Document, *types.Array, types.Binary,
		types.NullType, types.Regex, types.Timestamp:
		// type not supported for pushdown

	case float64:
		// If value is not safe double, fetch all numbers out of safe range.
		switch {
		case v > types.MaxSafeDouble:
			sql = `_jsonb->%[1]s > %[2]s`
			v = types.MaxSafeDouble

		case v < -types.MaxSafeDouble:
			sql = `_jsonb->%[1]s < %[2]s`
			v = -types.MaxSafeDouble
		default:
			// don't change the default eq query
		}

		filter = fmt.Sprintf(sql, p.Next(), p.Next())
		args = append(args, k, v)

	case string, types.ObjectID, time.Time:
		// don't change the default eq query
		filter = fmt.Sprintf(sql, p.Next(), p.Next())
		args = append(args, k, string(must.NotFail(sjson.MarshalSingleValue(v))))

	case bool, int32:
		// don't change the default eq query
		filter = fmt.Sprintf(sql, p.Next(), p.Next())
		args = append(args, k, v)

	case int64:
		maxSafeDouble := int64(types.MaxSafeDouble)

		// If value cannot be safe double, fetch all numbers out of the safe range.
		switch {
		case v > maxSafeDouble:
			sql = `_jsonb->%[1]s > %[2]s`
			v = maxSafeDouble

		case v < -maxSafeDouble:
			sql = `_jsonb->%[1]s < %[2]s`
			v = -maxSafeDouble
		default:
			// don't change the default eq query
		}

		filter = fmt.Sprintf(sql, p.Next(), p.Next())
		args = append(args, k, v)

	default:
		panic(fmt.Sprintf("Unexpected type of value: %v", v))
	}

	return
}
