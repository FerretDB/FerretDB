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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

func projection(projection *types.Document, p *pg.Placeholder) (sql string, args []any, err error) {
	if projection == nil {
		sql = "_jsonb"
		return
	}
	projectionMap := projection.Map()
	if len(projectionMap) == 0 {
		sql = "_jsonb"
		return
	}

	// create a list of keys for document
	ks, arg, err := buildProjectionKeys(projection.Keys(), projectionMap, p)
	if err != nil {
		err = lazyerrors.Errorf("buildProjectionKeys: %w", err)
		return
	}
	args = append(args, arg...)
	sql = "json_build_object('$k', array[" + ks + "], "

	// _id and _id value
	sql += p.Next() + "::text, _jsonb->" + p.Next() + " "
	args = append(args, "_id", "_id")

	// build json object
	for _, k := range projection.Keys() { // value
		if _, isDoc := projectionMap[k].(*types.Document); !isDoc {
			sql += ", "
			sql += p.Next() + "::text, _jsonb->" + p.Next()
			args = append(args, k, k)
			continue
		}
	}
	sql += ")"

	return
}

// buildProjectionKeys prepares a key list with placeholders.
func buildProjectionKeys(projectionKeys []string, projectionMap map[string]any, p *pg.Placeholder) (
	ks string, arg []any, err error,
) {
	ks += p.Next()
	arg = append(arg, "_id")

	for _, k := range projectionKeys {
		if _, isDoc := projectionMap[k].(*types.Document); !isDoc {
			ks += ", "
			ks += p.Next()
			arg = append(arg, k)
			continue
		}
	}
	return
}
