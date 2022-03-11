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

// elemMatch
// SELECT json_build_object('$k', array['value'], 'value'::text, _jsonb->$3) FROM "values"."values" WHERE (_jsonb->'name' = to_jsonb('array-embedded'::text))
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
			if elemMatchAny, err := doc.Get("$elemMatch"); err == nil {
				elemMatchDoc, ok := elemMatchAny.(*types.Document)
				if !ok {
					panic("expected $elemMatch to be doc")
				}
				for _, filterField := range elemMatchDoc.Keys() {
					if i != 0 {
						ks += ", "
					}
					ks += "" + p.Next()
					args = append(args, k)
					s.l.Sugar().Debugf("$elemMatch field [%s] in %s", filterField, k)
				}
				continue
			}
		}
		if i != 0 {
			ks += ", "
		}
		ks += p.Next()
		args = append(args, k)
	}

	sql = "json_build_object('$k', array[" + ks + "],"
	for i, k := range projection.Keys() { // value
		if i != 0 {
			sql += ", "
		}
		sql += p.Next() + "::text, _jsonb->" + p.Next()
		args = append(args, k, k)
	}
	sql += ")"

	return
}
