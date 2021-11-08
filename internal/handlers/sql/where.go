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

package sql

import (
	"github.com/jackc/pgx/v4"

	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

func where(d types.Document, placeholder *pg.Placeholder) (sql string, args []interface{}, err error) {
	filterMap := d.Map()
	if len(filterMap) > 0 {
		sql += " WHERE"

		for i, k := range d.Keys() {
			if i != 0 {
				sql += " AND"
			}

			sql += " " + pgx.Identifier{k}.Sanitize()
			v := filterMap[k]
			switch v.(type) {
			case int32:
				sql += " = " + placeholder.Next()
				args = append(args, v)
			case string:
				sql += " = " + placeholder.Next()
				args = append(args, v)
			default:
				err = lazyerrors.Errorf("unhandled field %s %v (%T)", k, v, v)
				return
			}
		}
	}

	return
}
