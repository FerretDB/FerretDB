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
	"github.com/MangoDB-io/MangoDB/internal/pgconn"
	"github.com/MangoDB-io/MangoDB/internal/types"
	lazyerrors "github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

func where(d types.Document, placeholder *pgconn.Placeholder) (sql string, args []interface{}, err error) {
	filterMap := d.Map()
	if len(filterMap) > 0 {
		sql += " WHERE"

		for i, k := range d.Keys() {
			if i != 0 {
				sql += " AND"
			}

			sql += " _jsonb->" + placeholder.Next()
			args = append(args, k)
			v := filterMap[k]
			switch v := v.(type) {
			case int32:
				sql += " = to_jsonb(" + placeholder.Next() + "::int4)"
				args = append(args, v)
			case string:
				sql += " = to_jsonb(" + placeholder.Next() + "::text)"
				args = append(args, v)
			case types.ObjectID:
				sql += " = " + placeholder.Next()
				var b []byte
				if b, err = bson.ObjectID(v).MarshalJSON(); err != nil {
					return
				}
				args = append(args, string(b))
			default:
				err = lazyerrors.Errorf("unhandled field %s %v (%T)", k, v, v)
				return
			}
		}
	}

	return
}
