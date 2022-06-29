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
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// FIXME sanitize input
func MatchToSql(field string, value interface{}, joinOp string, values *[]interface{}) string {
	var sql string

	fmt.Printf("  *** field, value: %v, %v, %+v\n", field, value, value)

	switch v := value.(type) {
	case *types.Document:
		sep := "->"
		if strings.Contains(field, "->") {
			sep = "->>"
		}
		for i, key := range v.Keys() {
			s := MatchToSql(field+sep+`'`+key+`'`, must.NotFail(v.Get(key)), joinOp, values)
			if i > 0 {
				sql += " " + joinOp + " "
			}
			sql += s
		}
	case *types.Array:
		if strings.HasSuffix(field, "'$or'") {
			if len(*values) > 0 {
				sql += " " + joinOp + " "
			}
			sql += "("
			for i := 0; i < v.Len(); i++ {
				name := strings.TrimSuffix(strings.TrimSuffix(field, "->'$or'"), "->>'$or'")
				if i > 0 {
					sql += " OR "
				}
				sql += MatchToSql(name, must.NotFail(v.Get(i)), "OR", values)
			}
			sql += ")"
		}

	default:
		*values = append(*values, fmt.Sprintf("%v", value))
		sql = field + ` = $` + fmt.Sprintf("%v", len(*values))
		fmt.Printf("  *** SQL [%v] = [%v]\n", value, sql)
	}

	return sql
}

func AggregateMatch(match *types.Document) (string, []interface{}) {
	values := make([]interface{}, 0)
	sql := MatchToSql("_jsonb", match, "AND", &values)

	return sql, values
}
