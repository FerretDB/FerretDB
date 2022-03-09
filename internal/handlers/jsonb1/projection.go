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

package jsonb1

import (
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
)

// Filter fields to return.
func (s *storage) projection(projection *types.Document, p *pg.Placeholder) (sql string, args []any, err error) {
	if projection == nil {
		sql = "_jsonb"
		return
	}
	projectionMap := projection.Map()
	if len(projectionMap) == 0 {
		sql = "_jsonb"
		return
	}

	ks := ""
	for i, k := range projection.Keys() {
		doc, isDoc := projectionMap[k].(*types.Document)
		if isDoc {
			if _, err := doc.Get("$elemMatch"); err == nil {
				s.l.Sugar().Debugf("filter projection %s", k)
				continue
			}
		}
		if i != 0 {
			ks += ", "
		}
		ks += p.Next()
		args = append(args, k)
	}

	if ks == "" {
		sql = "_jsonb"
		return
	}

	sql = "json_build_object('$k', array[" + ks + "],"
	for i, k := range projection.Keys() { // value

		doc, isDoc := projectionMap[k].(*types.Document)
		if isDoc {
			if _, err := doc.Get("$elemMatch"); err == nil {
				s.l.Sugar().Debugf("filter projection %s", k)

				// // {field: {$elemMatch: value}}
				// elemMatchMap := value.(*types.Document).Map()
				// for elemMatchField, elemMatchVal := range elemMatchMap {
				// 	if argSql != "" {
				// 		argSql += " AND"
				// 	}

				// 	elemMatchCondition, isDoc := elemMatchMap[elemMatchField].(*types.Document)

				// 	// {field1: {$elemMatch: { field2: value}}}
				// 	// SELECT _jsonb FROM "values"."values" WHERE (_jsonb->'value' @?  '$.score[*] ? (@ == 24 )'  )
				// 	if !isDoc {
				// 		argSql += fmt.Sprintf(" _jsonb->%[1]s @? '$.%[2]s[*] ? (@ == %[3]v)' ",
				// 			p.Next(), elemMatchField, elemMatchVal,
				// 		)
				// 		arg = append(arg, field)
				// 		continue
				// 	}

				// 	// {field1: { $elemMatch: { field2: { $gt: 23 }}} }
				// 	filterMap := elemMatchCondition.Map()
				// 	for elemMatchOp, val := range filterMap {
				// 		var operand string
				// 		switch elemMatchOp {
				// 		case "$eq":
				// 			operand = "=="
				// 		case "$ne":
				// 			operand = "<>"
				// 		case "$lt":
				// 			operand = "<"
				// 		case "$lte":
				// 			operand = "<="
				// 		case "$gt":
				// 			operand = ">"
				// 		case "$gte":
				// 			operand = ">="
				// 		}

				// 		argSql += fmt.Sprintf(" _jsonb->%[1]s @? '$.%[2]s[*] ? (@ %[3]s %[4]v)' ",
				// 			p.Next(), elemMatchField, operand, val,
				// 		)
				// 		arg = append(arg, field)
				// 	}
				// }
				continue
			}
		}

		if i != 0 {
			sql += ", "
		}
		sql += p.Next() + "::text, _jsonb->" + p.Next()
		args = append(args, k, k)
	}
	sql += ")"

	return
}
