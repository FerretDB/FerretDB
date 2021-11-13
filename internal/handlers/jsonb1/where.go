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

func scalar(v interface{}, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
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
		if argSql, arg, err = scalar(el, placeholder); err != nil {
			err = lazyerrors.Errorf("inArray: %w", err)
			return
		}
		sql += argSql
		args = append(args, arg...)
	}
	sql += ")"
	return
}

// fieldExpr handles {field: {expr}}.
func fieldExpr(field string, expr types.Document, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	filterKeys := expr.Keys()
	filterMap := expr.Map()

	sql = "("
	for i, op := range filterKeys {
		if i != 0 {
			sql += " AND"
		}

		// special case
		if op != "$not" {
			sql += " _jsonb->" + placeholder.Next()
			args = append(args, field)
		}

		value := filterMap[op]

		var argSql string
		var arg []interface{}
		switch op {
		case "$not":
			// {field: {$not: {expr}}}
			sql += " NOT"
			argSql, arg, err = fieldExpr(field, value.(types.Document), placeholder)
		case "$in":
			// {field: {$in: [value1, value2, ...]}}
			sql += " IN"
			argSql, arg, err = inArray(value.(types.Array), placeholder)
		case "$nin":
			// {field: {$nin: [value1, value2, ...]}}
			sql += " NOT IN"
			argSql, arg, err = inArray(value.(types.Array), placeholder)
		case "$eq":
			// {field: {$eq: value}}
			// TODO special handling for regex
			sql += " ="
			argSql, arg, err = scalar(value, placeholder)
		case "$ne":
			// {field: {$ne: value}}
			sql += " <>"
			argSql, arg, err = scalar(value, placeholder)
		case "$lt":
			// {field: {$lt: value}}
			sql += " <"
			argSql, arg, err = scalar(value, placeholder)
		case "$lte":
			// {field: {$lte: value}}
			sql += " <="
			argSql, arg, err = scalar(value, placeholder)
		case "$gt":
			// {field: {$gt: value}}
			sql += " >"
			argSql, arg, err = scalar(value, placeholder)
		case "$gte":
			// {field: {$gte: value}}
			sql += " >="
			argSql, arg, err = scalar(value, placeholder)
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

func wherePair(key string, value interface{}, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	switch value := value.(type) {
	case types.Document:
		// {field: {expr}}
		sql, args, err = fieldExpr(key, value, placeholder)

	default:
		// {field: value}
		sql = "_jsonb->" + placeholder.Next() + " = "
		args = append(args, key)

		var scalarSQL string
		var scalarArgs []interface{}
		scalarSQL, scalarArgs, err = scalar(value, placeholder)
		sql += scalarSQL
		args = append(args, scalarArgs...)
	}

	if err != nil {
		err = lazyerrors.Errorf("wherePair: %w", err)
	}

	return
}

func where(filter types.Document, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	filterMap := filter.Map()
	if len(filterMap) == 0 {
		return
	}

	sql = " WHERE"

	for i, key := range filter.Keys() {
		value := filterMap[key]

		if i != 0 {
			sql += " AND"
		}

		var argSql string
		var arg []interface{}
		argSql, arg, err = wherePair(key, value, placeholder)
		if err != nil {
			err = lazyerrors.Errorf("where: %w", err)
			return
		}

		sql += " (" + argSql + ")"
		args = append(args, arg...)
	}

	return
}
