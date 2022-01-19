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

package common

import (
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type wherePair func(key string, value any, p *pg.Placeholder) (sql string, args []any, err error)

func LogicExpr(op string, exprs *types.Array, p *pg.Placeholder, wherePair wherePair) (sql string, args []any, err error) {
	if op == "$nor" {
		sql = "NOT ("
	}

	switch op {
	case "$or", "$and", "$nor":
		// {$or: [{expr1}, {expr2}, ...]}
		// {$and: [{expr1}, {expr2}, ...]}
		// {$nor: [{expr1}, {expr2}, ...]}
		for i := 0; i < exprs.Len(); i++ {
			if i != 0 {
				switch op {
				case "$or", "$nor":
					sql += " OR"
				case "$and":
					sql += " AND"
				}
			}

			var el any
			if el, err = exprs.Get(i); err != nil {
				err = lazyerrors.Errorf("logicExpr: %w", err)
				return
			}

			expr := el.(*types.Document)
			m := expr.Map()
			for j, key := range expr.Keys() {
				if j != 0 {
					sql += " AND"
				}

				var exprSQL string
				var exprArgs []any
				exprSQL, exprArgs, err = wherePair(key, m[key], p)
				if err != nil {
					err = lazyerrors.Errorf("logicExpr: %w", err)
					return
				}

				if sql != "" {
					sql += " "
				}
				sql += "(" + exprSQL + ")"
				args = append(args, exprArgs...)
			}
		}

	default:
		err = lazyerrors.Errorf("logicExpr: unhandled op %q", op)
	}

	if op == "$nor" {
		sql += ")"
	}

	return
}

type scalar func(v any, p *pg.Placeholder) (sql string, args []any, err error)

func InArray(a *types.Array, p *pg.Placeholder, scalar scalar) (sql string, args []any, err error) {
	sql = "("
	for i := 0; i < a.Len(); i++ {
		if i != 0 {
			sql += ", "
		}

		var el any
		if el, err = a.Get(i); err != nil {
			err = lazyerrors.Errorf("inArray: %w", err)
			return
		}

		var argSql string
		var arg []any
		if argSql, arg, err = scalar(el, p); err != nil {
			err = lazyerrors.Errorf("inArray: %w", err)
			return
		}
		sql += argSql
		args = append(args, arg...)
	}
	sql += ")"
	return
}
