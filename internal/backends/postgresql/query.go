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

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// prepareSelectClause returns SELECT clause for default column of provided schema and table name.
//
// For capped collection with onlyRecordIDs, it returns select clause for recordID column.
//
// For capped collection, it returns select clause for recordID column and default column.
func prepareSelectClause(schema, table string, capped, onlyRecordIDs bool) string {
	if capped && onlyRecordIDs {
		return fmt.Sprintf(
			`SELECT %s FROM %s`,
			metadata.RecordIDColumn,
			pgx.Identifier{schema, table}.Sanitize(),
		)
	}

	if capped {
		return fmt.Sprintf(
			`SELECT %s, %s FROM %s`,
			metadata.RecordIDColumn,
			metadata.DefaultColumn,
			pgx.Identifier{schema, table}.Sanitize(),
		)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3573
	return fmt.Sprintf(
		`SELECT %s FROM %s`,
		metadata.DefaultColumn,
		pgx.Identifier{schema, table}.Sanitize(),
	)
}

// prepareWhereClause adds WHERE clause with given filters to the query and returns the query and arguments.
// It returns false if pushdown for at least one filter couldn't be applied.
func prepareWhereClause(p *metadata.Placeholder, sqlFilters *types.Document) (string, []any, bool, error) {
	var filters []string
	var args []any
	var filterPushdownAll bool

	filterPushdownAll = true

	iter := sqlFilters.Iterator()
	defer iter.Close()

	// iterate through root document
	for {
		rootKey, rootVal, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return "", nil, filterPushdownAll, lazyerrors.Error(err)
		}

		// Is the comment below correct? Does it also skip things like $or?
		// TODO https://github.com/FerretDB/FerretDB/issues/3573

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
				filterPushdownAll = false
				continue
			}
		case errors.As(err, &pe):
			// ignore empty key error, otherwise return error
			if pe.Code() != types.ErrPathElementEmpty {
				return "", nil, filterPushdownAll, lazyerrors.Error(err)
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

					return "", nil, filterPushdownAll, lazyerrors.Error(err)
				}

				switch k {
				case "$eq":
					if f, a, e := filterEqual(p, rootKey, v); f != "" {
						filters = append(filters, f)
						args = append(args, a...)
						// if equal filter is best-effort match, filter pushdown is not applied entirely
						if !e {
							filterPushdownAll = false
						}
					} else {
						filterPushdownAll = false
					}

				case "$ne":
					// $ne filter pushdown is best-effort match, filter pushdown is not applied entirely
					filterPushdownAll = false
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

						args = append(args, rootKey, v)

					case string, types.ObjectID, time.Time:
						filters = append(filters, fmt.Sprintf(
							sql,
							metadata.DefaultColumn,
							p.Next(),
							p.Next(),
							sjson.GetTypeOfValue(v),
						))

						args = append(args, rootKey, string(must.NotFail(sjson.MarshalSingleValue(v))))

					default:
						panic(fmt.Sprintf("Unexpected type of value: %v", v))
					}

				default:
					// $gt and $lt
					// TODO https://github.com/FerretDB/FerretDB/issues/1875
					filterPushdownAll = false
					continue
				}
			}

		case *types.Array, types.Binary, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown
			filterPushdownAll = false

		case float64, string, types.ObjectID, bool, time.Time, int32, int64:
			if f, a, e := filterEqual(p, rootKey, v); f != "" {
				filters = append(filters, f)
				args = append(args, a...)
				// if equal filter is best-effort match, filter pushdown is not considered as applied entirely
				if !e {
					filterPushdownAll = false
				}
			} else {
				filterPushdownAll = false
			}

		default:
			panic(fmt.Sprintf("Unexpected type of value: %v", v))
		}
	}

	var filter string
	if len(filters) > 0 {
		filter = ` WHERE ` + strings.Join(filters, " AND ")
	}

	return filter, args, filterPushdownAll, nil
}

// prepareOrderByClause returns ORDER BY clause for given sort field and returns the query and arguments.
// It returns false if pushdown for provided field couldn't be applied.
// Otherwise, if pushdown was applied, or field is nil, it returns true.
// For capped collection, it returns ORDER BY recordID only if sort field is nil.
func prepareOrderByClause(p *metadata.Placeholder, sort *backends.SortField, capped bool) (string, []any, bool) {
	if sort == nil {
		if capped {
			return fmt.Sprintf(" ORDER BY %s", metadata.RecordIDColumn), nil, true
		}

		return "", nil, true
	}

	// Skip sorting dot notation
	if strings.ContainsRune(sort.Key, '.') {
		return "", nil, false
	}

	var order string
	if sort.Descending {
		order = " DESC"
	}

	return fmt.Sprintf(" ORDER BY %s->%s%s", metadata.DefaultColumn, p.Next(), order), []any{sort.Key}, true
}

// filterEqual returns the proper SQL filter with arguments that filters documents
// where the value under k is equal to v, exactMatch returns true if the filter is exact match,
// returns false if it is best-effort match.
func filterEqual(p *metadata.Placeholder, k string, v any) (filter string, args []any, exactMatch bool) {
	// Select if value under the key is equal to provided value.
	sql := `%[1]s->%[2]s @> %[3]s`
	exactMatch = false

	switch v := v.(type) {
	case *types.Document, *types.Array, types.Binary,
		types.NullType, types.Regex, types.Timestamp:
		// type not supported for pushdown

	case float64:
		// If value is not safe double, fetch all numbers out of safe range.
		switch {
		case v > types.MaxSafeDouble:
			sql = `%[1]s->%[2]s > %[3]s`
			v = types.MaxSafeDouble

		case v < -types.MaxSafeDouble:
			sql = `%[1]s->%[2]s < %[3]s`
			v = -types.MaxSafeDouble
		default:
			// don't change the default eq query
			exactMatch = true
		}

		filter = fmt.Sprintf(sql, metadata.DefaultColumn, p.Next(), p.Next())
		args = append(args, k, v)

	case string, types.ObjectID, time.Time:
		// don't change the default eq query
		exactMatch = true
		filter = fmt.Sprintf(sql, metadata.DefaultColumn, p.Next(), p.Next())
		args = append(args, k, string(must.NotFail(sjson.MarshalSingleValue(v))))

	case bool, int32:
		// don't change the default eq query
		exactMatch = true
		filter = fmt.Sprintf(sql, metadata.DefaultColumn, p.Next(), p.Next())
		args = append(args, k, v)

	case int64:
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
			exactMatch = true
		}

		filter = fmt.Sprintf(sql, metadata.DefaultColumn, p.Next(), p.Next())
		args = append(args, k, v)

	default:
		panic(fmt.Sprintf("Unexpected type of value: %v", v))
	}

	return
}
