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

package sqlite

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// prepareWhereClause adds WHERE clause with filters given in the document.
// It returns the WHERE clause and the SQLite arguments.
func prepareWhereClause(filterDoc *types.Document) (string, []any, error) {
	if filterDoc == nil {
		return "", []any{}, nil
	}

	iter := filterDoc.Iterator()
	defer iter.Close()

	var filters []string
	var args []any

	for {
		k, v, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return "", nil, lazyerrors.Error(err)
		}

		// queryPath stores the path that is used in SQLite to access specific key
		// if the key is _id we use our predifined path, as the handling of _id may
		// change in the future
		queryPath := metadata.IDColumn

		// keyArgs store the optional parameters used to query the key
		var keyArgs []any

		if k != "_id" {
			// To use parameters inside of SQLite json path the parameter token ("?")
			// needs to be concatenated to path with || operator
			queryPath = fmt.Sprintf(`%s->('$."' || ? || '"' )`, metadata.DefaultColumn)
			keyArgs = append(keyArgs, k)
		}

		// don't pushdown $comment
		if strings.HasPrefix(k, "$") {
			continue
		}

		path, err := types.NewPathFromString(k)

		var pe *types.PathError

		switch {
		case err == nil:
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

		switch v := v.(type) {
		case *types.Document, *types.Array, types.Binary, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown
			continue

		case float64:
			comparison := ` = ?`

			switch {
			case v > types.MaxSafeDouble:
				comparison = ` > ?`
				v = types.MaxSafeDouble

			case v < -types.MaxSafeDouble:
				comparison = ` < ?`
				v = -types.MaxSafeDouble
			default:
				// don't change the default eq query
			}

			subquery := fmt.Sprintf(`EXISTS (SELECT value FROM json_each(%s) WHERE value %s)`, queryPath, comparison)
			filters = append(filters, subquery)

			// TODO https://github.com/FerretDB/FerretDB/issues/3386
			args = append(args, keyArgs...)
			args = append(args, parseValue(v))

		case types.ObjectID, time.Time, string, bool, int32:
			subquery := fmt.Sprintf(`EXISTS (SELECT value FROM json_each(%s) WHERE value = ?)`, queryPath)
			filters = append(filters, subquery)

			// TODO https://github.com/FerretDB/FerretDB/issues/3386
			args = append(args, keyArgs...)
			args = append(args, parseValue(v))

		case int64:
			comparison := ` = ?`
			maxSafeDouble := int64(types.MaxSafeDouble)

			// If value cannot be safe double, fetch all numbers out of the safe range
			switch {
			case v > maxSafeDouble:
				comparison = ` > ?`
				v = maxSafeDouble

			case v < -maxSafeDouble:
				comparison = `< ?`
				v = -maxSafeDouble
			default:
				// don't change the default eq query
			}

			// json_each returns top level json values, and the contents of arrays if any
			// https://www.sqlite.org/json1.html#jeach
			subquery := fmt.Sprintf(`EXISTS (SELECT value FROM json_each(%s) WHERE value %s)`, queryPath, comparison)
			filters = append(filters, subquery)

			// TODO https://github.com/FerretDB/FerretDB/issues/3386
			args = append(args, keyArgs...)
			args = append(args, parseValue(v))

		default:
			panic(fmt.Sprintf("Unexpected type of value: %v", v))
		}
	}

	var whereClause string
	if len(filters) > 0 {
		whereClause = ` WHERE ` + strings.Join(filters, " AND ")
	}

	return whereClause, args, nil
}
