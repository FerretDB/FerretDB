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

// GetPushdownQuery gets pushdown query ($match and $sort) for aggregation.
//
// If the first two stages are either $match, $sort, or a combination of them, we can push them down.
// In this case, we return the first match and sort statements to pushdown.
// If $match stage is not present, match is returned as nil.
// If $sort stage is not present, sort is returned as nil.
func GetPushdownQuery(stagesDocs []any) (match *types.Document, sort *types.Document) {
	if len(stagesDocs) == 0 {
		return
	}

	stagesToPushdown := []any{stagesDocs[0]}

	if len(stagesDocs) > 1 {
		stagesToPushdown = append(stagesToPushdown, stagesDocs[1])
	}

	for _, s := range stagesToPushdown {
		stage, isDoc := s.(*types.Document)

		if !isDoc {
			return nil, nil
		}

		switch {
		case stage.Has("$match"):
			matchQuery := must.NotFail(stage.Get("$match"))
			query, isDoc := matchQuery.(*types.Document)

			if !isDoc || match != nil {
				continue
			}

			match = query

		case stage.Has("$sort"):
			sortQuery := must.NotFail(stage.Get("$sort"))
			query, isDoc := sortQuery.(*types.Document)

			if !isDoc || sort != nil {
				continue
			}

			sort = query

		default:
			// not $match nor $sort, we shouldn't continue pushdown
			return
		}
	}

	return
}
