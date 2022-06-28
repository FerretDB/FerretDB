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
func GetValue(field string, value interface{}) ([]string, []interface{}) {
	var fields []string
	var values []interface{}

	switch v := value.(type) {
	case *types.Document:
		sep := "->"
		if strings.Contains(field, "->") {
			sep = "->>"
		}
		for _, key := range v.Keys() {
			f, v := GetValue(field+sep+`'`+key+`'`, must.NotFail(v.Get(key)))
			fields = append(fields, f...)
			values = append(values, v...)
		}
	default:
		fields = append(fields, field)
		values = append(values, fmt.Sprintf("%v", value))
	}

	return fields, values
}

func AggregateMatch(match *types.Document) (string, []interface{}) {
	var where []string

	sql := ``
	if len(match.Keys()) > 0 {
		sql += ` WHERE`
	}

	fields, values := GetValue("_jsonb", match)
	for i, field := range fields {
		where = append(where, fmt.Sprintf("%s = $%v", field, i+1))
	}

	sql += ` ` + strings.Join(where, ` AND `)
	return sql, values
}
