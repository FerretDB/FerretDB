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

package sql

import (
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

func scalar(v any, p *pg.Placeholder) (sql string, args []any, err error) {
	sql = p.Next()

	switch v := v.(type) {
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
		s := v.Pattern
		if options != "" {
			s = "(?" + options + ")" + v.Pattern
		}
		args = []any{s}
	default:
		args = []any{v}
	}
	return
}

// fieldExpr handles {field: {expr}}.
func fieldExpr(field string, expr *types.Document, p *pg.Placeholder) (sql string, args []any, err error) {
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
		var arg []any
		value := filterMap[op]

		// {field: {$not: {expr}}}
		if op == "$not" {
			if sql != "" {
				sql += " "
			}
			sql += "NOT("

			argSql, arg, err = fieldExpr(field, value.(*types.Document), p)
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
		sql += pgx.Identifier{field}.Sanitize()

		switch op {
		case "$in":
			// {field: {$in: [value1, value2, ...]}}
			sql += " IN"
			argSql, arg, err = common.InArray(value.(*types.Array), p, scalar)
		case "$nin":
			// {field: {$nin: [value1, value2, ...]}}
			sql += " NOT IN"
			argSql, arg, err = common.InArray(value.(*types.Array), p, scalar)
		case "$eq":
			// {field: {$eq: value}}
			// TODO special handling for regex
			sql += " ="
			argSql, arg, err = scalar(value, p)
		case "$ne":
			// {field: {$ne: value}}
			sql += " <>"
			argSql, arg, err = scalar(value, p)
		case "$lt":
			// {field: {$lt: value}}
			sql += " <"
			argSql, arg, err = scalar(value, p)
		case "$lte":
			// {field: {$lte: value}}
			sql += " <="
			argSql, arg, err = scalar(value, p)
		case "$gt":
			// {field: {$gt: value}}
			sql += " >"
			argSql, arg, err = scalar(value, p)
		case "$gte":
			// {field: {$gte: value}}
			sql += " >="
			argSql, arg, err = scalar(value, p)
		case "$regex":
			// {field: {$regex: value}}

			var options string
			if opts, ok := filterMap["$options"]; ok {
				// {field: {$regex: value, $options: string}}
				if options, ok = opts.(string); !ok {
					err = common.NewErrorMsg(common.ErrBadValue, "$options has to be a string")
					return
				}
			}

			sql += " ~"
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
						err = common.NewErrorMsg(common.ErrRegexOptions, "options set in both $regex and $options")
						return
					}
					value.Options = options
				}
				argSql, arg, err = scalar(value, p)
			default:
				err = common.NewErrorMsg(common.ErrBadValue, "$regex has to be a string")
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

func wherePair(key string, value any, p *pg.Placeholder) (sql string, args []any, err error) {
	if strings.HasPrefix(key, "$") {
		exprs := value.(*types.Array)
		sql, args, err = common.LogicExpr(key, exprs, p, wherePair)
		return
	}

	switch value := value.(type) {
	case *types.Document:
		// {field: {expr}}
		sql, args, err = fieldExpr(key, value, p)

	default:
		// {field: value}
		sql, args, err = scalar(value, p)
		switch value.(type) {
		case types.Regex:
			sql = pgx.Identifier{key}.Sanitize() + " ~ " + sql
		default:
			sql = pgx.Identifier{key}.Sanitize() + " = " + sql
		}
	}

	if err != nil {
		err = lazyerrors.Errorf("wherePair: %w", err)
	}

	return
}

func where(filter *types.Document, p *pg.Placeholder) (sql string, args []any, err error) {
	if filter == nil {
		return
	}
	filterMap := filter.Map()
	if len(filterMap) == 0 {
		return
	}

	sql += " WHERE"

	for i, key := range filter.Keys() {
		value := filterMap[key]

		if i != 0 {
			sql += " AND"
		}

		var argSql string
		var arg []any
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
