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

package jsonb1

import (
	"github.com/MangoDB-io/MangoDB/internal/bson"
	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

func scalarValue(v interface{}, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	var arg interface{}
	switch v := v.(type) {
	case int32:
		sql = "to_jsonb(" + placeholder.Next() + "::int4)"
		arg = v
	case string:
		sql = "to_jsonb(" + placeholder.Next() + "::text)"
		arg = v
	case types.ObjectID:
		sql = placeholder.Next()
		var b []byte
		if b, err = bson.ObjectID(v).MarshalJSON(); err != nil {
			err = lazyerrors.Errorf("jsonArgument: %w", err)
			return
		}
		arg = string(b)
	default:
		err = lazyerrors.Errorf("jsonArgument: unhandled field %v (%T)", v, v)
	}

	args = []interface{}{arg}
	return
}

func inArray(a types.Array, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	sql = "("
	for i, el := range a {
		if i != 0 {
			sql += ", "
		}

		var argSql string
		var arg []interface{}
		if argSql, arg, err = scalarValue(el, placeholder); err != nil {
			err = lazyerrors.Errorf("inArray: %w", err)
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
	for i, op := range filterKeys {
		if i != 0 {
			sql += " AND"
		}

		sql += " _jsonb->" + placeholder.Next()
		args = append(args, field)
		value := filterMap[op]

		var argSql string
		var arg []interface{}
		switch op {
		case "$in":
			sql += " IN"
			argSql, arg, err = inArray(value.(types.Array), placeholder)
		case "$nin":
			sql += " NOT IN"
			argSql, arg, err = inArray(value.(types.Array), placeholder)
		case "$eq":
			sql += " ="
			argSql, arg, err = scalarValue(value, placeholder)
		case "$ne":
			sql += " <>"
			argSql, arg, err = scalarValue(value, placeholder)
		case "$lt":
			sql += " <"
			argSql, arg, err = scalarValue(value, placeholder)
		case "$lte":
			sql += " <="
			argSql, arg, err = scalarValue(value, placeholder)
		case "$gt":
			sql += " >"
			argSql, arg, err = scalarValue(value, placeholder)
		case "$gte":
			sql += " >="
			argSql, arg, err = scalarValue(value, placeholder)
		default:
			err = lazyerrors.Errorf("unhandled {%q: %v}", op, value)
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
			sql += " _jsonb->" + placeholder.Next() + " ="
			args = append(args, filterKey)
			argSql, arg, err = scalarValue(filterValue, placeholder)
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
