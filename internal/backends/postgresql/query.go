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

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
	"github.com/FerretDB/FerretDB/internal/handler/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// selectParams contains params that specify how prepareSelectClause function will
// build the SELECT SQL query.
type selectParams struct {
	Schema  string
	Table   string
	Comment string

	Capped        bool
	OnlyRecordIDs bool
}

// prepareSelectClause returns SELECT clause for default column of provided schema and table name.
//
// For capped collection with onlyRecordIDs, it returns select clause for recordID column.
//
// For capped collection, it returns select clause for recordID column and default column.
func prepareSelectClause(params *selectParams) string {
	if params == nil {
		params = new(selectParams)
	}

	if params.Comment != "" {
		params.Comment = strings.ReplaceAll(params.Comment, "/*", "/ *")
		params.Comment = strings.ReplaceAll(params.Comment, "*/", "* /")
		params.Comment = `/* ` + params.Comment + ` */`
	}

	if params.Capped && params.OnlyRecordIDs {
		return fmt.Sprintf(
			`SELECT %s %s FROM %s`,
			params.Comment,
			metadata.RecordIDColumn,
			pgx.Identifier{params.Schema, params.Table}.Sanitize(),
		)
	}

	if params.Capped {
		return fmt.Sprintf(
			`SELECT %s %s, %s FROM %s`,
			params.Comment,
			metadata.RecordIDColumn,
			metadata.DefaultColumn,
			pgx.Identifier{params.Schema, params.Table}.Sanitize(),
		)
	}

	return fmt.Sprintf(
		`SELECT %s %s FROM %s`,
		params.Comment,
		metadata.DefaultColumn,
		pgx.Identifier{params.Schema, params.Table}.Sanitize(),
	)
}

// prepareWhereClause adds WHERE clause with given filters to the query and returns the query and arguments.
func prepareWhereClause(p *metadata.Placeholder, sqlFilters *types.Document) (string, []any, error) {
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
				return "", nil, lazyerrors.Error(err)
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

					return "", nil, lazyerrors.Error(err)
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
						`%[1]s ? %[2]s AND ` +
						// does the value under the key is equal to filter value
						`%[1]s->%[2]s @> %[3]s AND ` +
						// does the value type is equal to the filter's one
						`%[1]s->'$s'->'p'->%[2]s->'t' = '"%[4]s"' )`

					switch v := v.(type) {
					case *types.Document, *types.Array, types.Binary,
						types.NullType, types.Regex, types.Timestamp:
						// type not supported for pushdown

					case float64, bool, int32, int64:
						filters = append(filters, fmt.Sprintf(
							sql,
							metadata.DefaultColumn,
							p.Next(),
							p.Next(),
							sjson.GetTypeOfValue(v),
						))

						// merge with the case below?
						// TODO https://github.com/FerretDB/FerretDB/issues/3626
						args = append(args, rootKey, v)

					case string, types.ObjectID, time.Time:
						filters = append(filters, fmt.Sprintf(
							sql,
							metadata.DefaultColumn,
							p.Next(),
							p.Next(),
							sjson.GetTypeOfValue(v),
						))

						// merge with the case above?
						// TODO https://github.com/FerretDB/FerretDB/issues/3626
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
		filter = ` WHERE ` + strings.Join(filters, " AND ")
	}

	return filter, args, nil
}

// prepareOrderByClause returns ORDER BY clause with arguments for given sort document.
//
// The provided sort document should be already validated.
//
// For more than one sort fields, it sorts only by the first key provided.
func prepareOrderByClause(p *metadata.Placeholder, sort *types.Document) (string, []any) {
	if sort.Len() == 0 {
		return "", nil
	}

	k := sort.Keys()[0]
	v := sort.Values()[0].(int64)

	if k == "$natural" {
		if v == 1 {
			return fmt.Sprintf(" ORDER BY %s", metadata.RecordIDColumn), nil
		}

		return "", nil
	}

	// Skip sorting dot notation
	if strings.ContainsRune(k, '.') {
		return "", nil
	}

	var order string
	if v == -1 {
		order = " DESC"
	}

	return fmt.Sprintf(" ORDER BY %s->%s%s", metadata.DefaultColumn, p.Next(), order), []any{k}
}

// filterEqual returns the proper SQL filter with arguments that filters documents
// where the value under k is equal to v.
func filterEqual(p *metadata.Placeholder, k string, v any) (filter string, args []any) {
	// Select if value under the key is equal to provided value.
	sql := `%[1]s->%[2]s @> %[3]s`

	switch v := v.(type) {
	case *types.Document, *types.Array, types.Binary,
		types.NullType, types.Regex, types.Timestamp:
		// type not supported for pushdown

	case float64:
		// If value is not safe double, fetch all numbers out of safe range.
		// TODO https://github.com/FerretDB/FerretDB/issues/3626
		switch {
		case v > types.MaxSafeDouble:
			sql = `%[1]s->%[2]s > %[3]s`
			v = types.MaxSafeDouble

		case v < -types.MaxSafeDouble:
			sql = `%[1]s->%[2]s < %[3]s`
			v = -types.MaxSafeDouble
		default:
			// don't change the default eq query
		}

		filter = fmt.Sprintf(sql, metadata.DefaultColumn, p.Next(), p.Next())
		args = append(args, k, v)

	case string, types.ObjectID, time.Time:
		// merge with the case below?
		// TODO https://github.com/FerretDB/FerretDB/issues/3626

		// don't change the default eq query
		filter = fmt.Sprintf(sql, metadata.DefaultColumn, p.Next(), p.Next())
		args = append(args, k, string(must.NotFail(sjson.MarshalSingleValue(v))))

	case bool, int32:
		// merge with the case above?
		// TODO https://github.com/FerretDB/FerretDB/issues/3626

		// don't change the default eq query
		filter = fmt.Sprintf(sql, metadata.DefaultColumn, p.Next(), p.Next())
		args = append(args, k, v)

	case int64:
		// TODO https://github.com/FerretDB/FerretDB/issues/3626
		maxSafeDouble := int64(types.MaxSafeDouble)

		// If value cannot be safe double, fetch all numbers out of the safe range.
		switch {
		case v > maxSafeDouble:
			sql = `%[1]s->%[2]s > %[3]s`
			v = maxSafeDouble

		case v < -maxSafeDouble:
			sql = `%[1]s->%[2]s < %[3]s`
			v = -maxSafeDouble
		default:
			// don't change the default eq query
		}

		filter = fmt.Sprintf(sql, metadata.DefaultColumn, p.Next(), p.Next())
		args = append(args, k, v)

	default:
		panic(fmt.Sprintf("Unexpected type of value: %v", v))
	}

	return
}
