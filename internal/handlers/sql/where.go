// Copyright 2021 Baltoro OÜ.
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

package sql

import (
	"github.com/jackc/pgx/v4"

	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

func sqlValue(v interface{}, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	sql = placeholder.Next()
	args = []interface{}{v}
	return
}

func array(a types.Array, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	sql = "("
	for i, el := range a {
		if i != 0 {
			sql += ", "
		}

		var argSql string
		var arg []interface{}
		if argSql, arg, err = sqlValue(el, placeholder); err != nil {
			err = lazyerrors.Errorf("array: %w", err)
			return
		}
		sql += argSql
		args = append(args, arg...)
	}
	sql += ")"
	return
}

func filterObject(field string, filter types.Document, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	filterKeys := filter.Keys()
	filterMap := filter.Map()

	sql = "("
	for i, filterKey := range filterKeys {
		if i != 0 {
			sql += " AND"
		}

		sql += " " + pgx.Identifier{field}.Sanitize()

		filterValue := filterMap[filterKey]

		var argSql string
		var arg []interface{}
		switch filterKey {
		case "$in":
			sql += " IN"
			argSql, arg, err = array(filterValue.(types.Array), placeholder)
		case "$nin":
			sql += " NOT IN"
			argSql, arg, err = array(filterValue.(types.Array), placeholder)
		case "$lt":
			sql += " <"
			argSql, arg, err = sqlValue(filterValue, placeholder)
		case "$gt":
			sql += " >"
			argSql, arg, err = sqlValue(filterValue, placeholder)
		default:
			err = lazyerrors.Errorf("unhandled {%q: %v}", filterKey, filterValue)
		}

		if err != nil {
			err = lazyerrors.Errorf("filterObject: %w", err)
			return
		}
		sql += " " + argSql
		args = append(args, arg...)
	}

	sql += ")"
	return
}

func where(filter types.Document, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	filterMap := filter.Map()
	if len(filterMap) == 0 {
		return
	}

	sql += " WHERE"

	for filterIndex, filterKey := range filter.Keys() {
		if filterIndex != 0 {
			sql += " AND"
		}

		filterValue := filterMap[filterKey]

		var argSql string
		var arg []interface{}
		switch filterValue := filterValue.(type) {
		case types.Document:
			argSql, arg, err = filterObject(filterKey, filterValue, placeholder)

		default:
			sql += " " + pgx.Identifier{filterKey}.Sanitize() + " ="
			argSql, arg, err = sqlValue(filterValue, placeholder)
		}

		if err != nil {
			err = lazyerrors.Errorf("where: %w", err)
			return
		}
		sql += " " + argSql
		args = append(args, arg...)
	}

	return
}
