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
func GetPushdownQuery(stagesDocs []any) *types.Document {
	if len(stagesDocs) == 0 {
		return nil
	}

	firstStageDoc := stagesDocs[0]
	firstStage, isDoc := firstStageDoc.(*types.Document)

	if !isDoc || !firstStage.Has("$match") {
		return nil
	}

	matchQuery := must.NotFail(firstStage.Get("$match"))
	if query, isDoc := matchQuery.(*types.Document); isDoc {
		return query
	}

	return nil
}
