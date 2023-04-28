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

package stages

import (
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// GetPushdownQuery gets pushdown query for aggregation.
// When the first aggregation stage is $match, $match query is
// used for pushdown, otherwise nil is return.
func GetPushdownQuery(stagesDocs []any) (*types.Document, *types.Document) {
	if len(stagesDocs) == 0 {
		return nil, nil
	}

	var match, sort *types.Document

	for i, doc := range stagesDocs {
		if i > 2 {
			break
		}

		stage, isDoc := doc.(*types.Document)
		if !isDoc {
			return nil, nil
		}

		switch {
		case stage.Has("$match"):
			matchQuery := must.NotFail(stage.Get("$match"))
			query, isDoc := matchQuery.(*types.Document)

			if !isDoc || match != nil {
				return nil, nil
			}

			match = query

		case stage.Has("$sort"):
			sortQuery := must.NotFail(stage.Get("$sort"))
			query, isDoc := sortQuery.(*types.Document)

			if !isDoc || sort != nil {
				return nil, nil
			}

			sort = query

		default:
			return nil, nil
		}
	}

	return match, sort
}
