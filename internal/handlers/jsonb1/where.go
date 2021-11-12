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
	"go.uber.org/zap"

	"github.com/MangoDB-io/MangoDB/internal/bson"
	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

func jsonValue(v interface{}, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
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

func array(a types.Array, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	sql = "("
	for i, el := range a {
		if i != 0 {
			sql += ", "
		}

		var argSql string
		var arg []interface{}
		if argSql, arg, err = jsonValue(el, placeholder); err != nil {
			err = lazyerrors.Errorf("array: %w", err)
			return
		}
		sql += argSql
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

		sql += " _jsonb->" + placeholder.Next()
		args = append(args, filterKey)

		filterValue := filterMap[filterKey]

		var argSql string
		var arg []interface{}
		switch filterValue := filterValue.(type) {
		case types.Document:
			keys := filterValue.Keys()
			if l := len(keys); l != 1 {
				err = lazyerrors.Errorf("unhandled field {%q: %v} (%d keys)", filterKey, filterValue, l)
				return
			}
			key := keys[0]
			value := filterValue.Map()[key]

			switch key {
			case "$in":
				sql += " IN "
				argSql, arg, err = array(value.(types.Array), placeholder)
			case "$nin":
				sql += " NOT IN "
				argSql, arg, err = array(value.(types.Array), placeholder)
			default:
				err = lazyerrors.Errorf("unhandled field {%q: {%q: %v}}", filterKey, filterValue, key)
				return
			}

		default:
			sql += " = "
			argSql, arg, err = jsonValue(filterValue, placeholder)
		}

		zap.S().Debugf("where: %v %v", argSql, arg)

		if err != nil {
			err = lazyerrors.Errorf("where: %w", err)
			return
		}
		sql += argSql
		args = append(args, arg...)
	}

	return
}
