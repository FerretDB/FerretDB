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
	"fmt"
	"strings"

	"github.com/MangoDB-io/MangoDB/internal/bson"
	"github.com/MangoDB-io/MangoDB/internal/handlers/common"
	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

func scalar(v interface{}, p *pg.Placeholder) (sql string, args []interface{}, err error) {
	var arg interface{}
	switch v := v.(type) {
	case int32:
		sql = "to_jsonb(" + p.Next() + "::int4)"
		arg = v
	case string:
		sql = "to_jsonb(" + p.Next() + "::text)"
		arg = v
	case types.ObjectID:
		sql = p.Next()
		var b []byte
		if b, err = bson.ObjectID(v).MarshalJSON(); err != nil {
			err = lazyerrors.Errorf("scalar: %w", err)
			return
		}
		arg = string(b)
	case types.Regex:
		var options string
		for _, o := range v.Options {
			switch o {
			case 'i':
				options += "i"
			default:
				err = lazyerrors.Errorf("scalar: unhandled regex option %v (%v)", o, v)
			}
		}
		sql = p.Next()
		arg = v.Pattern
		if options != "" {
			arg = "(?" + options + ")" + v.Pattern
		}
	default:
		err = lazyerrors.Errorf("scalar: unhandled field %v (%T)", v, v)
	}

	args = []interface{}{arg}
	return
}

// fieldExpr handles {field: {expr}}.
func fieldExpr(field string, expr types.Document, p *pg.Placeholder) (sql string, args []interface{}, err error) {
	filterKeys := expr.Keys()
	filterMap := expr.Map()

	for _, op := range filterKeys {
		if op == "$options" {
			// handled by $regex, no need to modify sql in any way
			continue
		}

		if sql != "" {
			sql += " AND"
		}

		var argSql string
		var arg []interface{}
		value := filterMap[op]

		// {field: {$not: {expr}}}
		if op == "$not" {
			if sql != "" {
				sql += " "
			}
			sql += "NOT("

			argSql, arg, err = fieldExpr(field, value.(types.Document), p)
			if err != nil {
				err = lazyerrors.Errorf("fieldExpr: %w", err)
				return
			}

			sql += argSql + ")"
			args = append(args, arg...)

			continue
		}

		if sql != "" {
			sql += " "
		}
		args = append(args, field)

		switch op {
		case "$in":
			// {field: {$in: [value1, value2, ...]}}
			sql += "_jsonb->" + p.Next() + " IN"
			argSql, arg, err = common.InArray(value.(types.Array), p, scalar)
		case "$nin":
			// {field: {$nin: [value1, value2, ...]}}
			sql += "_jsonb->" + p.Next() + " NOT IN"
			argSql, arg, err = common.InArray(value.(types.Array), p, scalar)
		case "$eq":
			// {field: {$eq: value}}
			// TODO special handling for regex
			sql += "_jsonb->" + p.Next() + " ="
			argSql, arg, err = scalar(value, p)
		case "$ne":
			// {field: {$ne: value}}
			sql += "_jsonb->" + p.Next() + " <>"
			argSql, arg, err = scalar(value, p)
		case "$lt":
			// {field: {$lt: value}}
			sql += "_jsonb->" + p.Next() + " <"
			argSql, arg, err = scalar(value, p)
		case "$lte":
			// {field: {$lte: value}}
			sql += "_jsonb->" + p.Next() + " <="
			argSql, arg, err = scalar(value, p)
		case "$gt":
			// {field: {$gt: value}}
			sql += "_jsonb->" + p.Next() + " >"
			argSql, arg, err = scalar(value, p)
		case "$gte":
			// {field: {$gte: value}}
			sql += "_jsonb->" + p.Next() + " >="
			argSql, arg, err = scalar(value, p)
		case "$regex":
			// {field: {$regex: value}}

			var options string
			if opts, ok := filterMap["$options"]; ok {
				// {field: {$regex: value, $options: string}}
				if options, ok = opts.(string); !ok {
					err = common.NewError(common.ErrBadValue, fmt.Errorf("$options has to be a string"))
					return
				}
			}

			sql += "_jsonb->>" + p.Next() + " ~"
			switch value := value.(type) {
			case string:
				// {field: {$regex: string}}
				v := types.Regex{
					Pattern: value,
					Options: options,
				}
				argSql, arg, err = scalar(v, p)
			case types.Regex:
				// {field: {$regex: /regex/}}
				if options != "" {
					if value.Options != "" {
						err = common.NewError(common.ErrRegexOptions, fmt.Errorf("options set in both $regex and $options"))
						return
					}
					value.Options = options
				}
				argSql, arg, err = scalar(value, p)
			default:
				err = common.NewError(common.ErrBadValue, fmt.Errorf("$regex has to be a string"))
				return
			}
		default:
			err = lazyerrors.Errorf("unhandled {%q: %v}", op, value)
		}

		if err != nil {
			err = lazyerrors.Errorf("fieldExpr: %w", err)
			return
		}

		sql += " " + argSql
		args = append(args, arg...)
	}

	return
}

func wherePair(key string, value interface{}, p *pg.Placeholder) (sql string, args []interface{}, err error) {
	if strings.HasPrefix(key, "$") {
		exprs := value.(types.Array)
		sql, args, err = common.LogicExpr(key, exprs, p, wherePair)
		return
	}

	switch value := value.(type) {
	case types.Document:
		// {field: {expr}}
		sql, args, err = fieldExpr(key, value, p)

	default:
		// {field: value}
		switch value.(type) {
		case types.Regex:
			sql = "_jsonb->>" + p.Next() + " ~ "
		default:
			sql = "_jsonb->" + p.Next() + " = "
		}

		args = append(args, key)

		var scalarSQL string
		var scalarArgs []interface{}
		scalarSQL, scalarArgs, err = scalar(value, p)
		sql += scalarSQL
		args = append(args, scalarArgs...)
	}

	if err != nil {
		err = lazyerrors.Errorf("wherePair: %w", err)
	}

	return
}

func where(filter types.Document, p *pg.Placeholder) (sql string, args []interface{}, err error) {
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
		argSql, arg, err = wherePair(key, value, p)
		if err != nil {
			err = lazyerrors.Errorf("where: %w", err)
			return
		}

		sql += " (" + argSql + ")"
		args = append(args, arg...)
	}

	return
}
