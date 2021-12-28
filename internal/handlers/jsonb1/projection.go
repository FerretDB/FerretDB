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

func projection(projection types.Document, p *pg.Placeholder) (sql string, args []any, err error) {
	projectionMap := projection.Map()
	if len(projectionMap) == 0 {
		sql = "_jsonb"
		return
	}

	ks := ""
	for i, k := range projection.Keys() {
		if i != 0 {
			ks += ", "
		}
		ks += p.Next()
		args = append(args, k)
	}
	sql = "json_build_object('$k', array[" + ks + "],"
	for i, k := range projection.Keys() {
		if i != 0 {
			sql += ", "
		}
		sql += p.Next() + "::text, _jsonb->" + p.Next()
		args = append(args, k, k)
	}
	sql += ")"

	return
}
