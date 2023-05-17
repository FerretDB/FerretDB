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

// Package stages provides aggregation stages.
package stages

import (
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
)

// newStageFunc is a type for a function that creates a new aggregation stage.
type newStageFunc func(stage *types.Document) (aggregations.Stage, error)

// stages maps all supported aggregation stages.
var stages = map[string]newStageFunc{
	// sorted alphabetically
	"$collStats": newCollStats,
	"$count":     newCount,
	"$group":     newGroup,
	"$limit":     newLimit,
	"$match":     newMatch,
	"$project":   newProject,
	"$skip":      newSkip,
	"$sort":      newSort,
	"$unwind":    newUnwind,
	// please keep sorted alphabetically
}

// unsupportedStages maps all unsupported yet stages.
var unsupportedStages = map[string]struct{}{
	// sorted alphabetically
	"$addFields":              {},
	"$bucket":                 {},
	"$bucketAuto":             {},
	"$changeStream":           {},
	"$currentOp":              {},
	"$densify":                {},
	"$documents":              {},
	"$facet":                  {},
	"$fill":                   {},
	"$geoNear":                {},
	"$graphLookup":            {},
	"$indexStats":             {},
	"$listLocalSessions":      {},
	"$listSessions":           {},
	"$lookup":                 {},
	"$merge":                  {},
	"$out":                    {},
	"$planCacheStats":         {},
	"$redact":                 {},
	"$replaceRoot":            {},
	"$replaceWith":            {},
	"$sample":                 {},
	"$search":                 {},
	"$searchMeta":             {},
	"$set":                    {},
	"$setWindowFields":        {},
	"$sharedDataDistribution": {},
	"$sortByCount":            {},
	"$unionWith":              {},
	"$unset":                  {},
	// please keep sorted alphabetically
}

// NewStage creates a new aggregation stage.
func NewStage(stage *types.Document) (aggregations.Stage, error) {
	if stage.Len() != 1 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageInvalid,
			"A pipeline stage specification object must contain exactly one field.",
			"aggregate",
		)
	}

	name := stage.Command()

	f, supported := stages[name]
	_, unsupported := unsupportedStages[name]

	switch {
	case supported && unsupported:
		panic(fmt.Sprintf("stage %q is in both `stages` and `unsupportedStages`", name))

	case supported && !unsupported:
		return f(stage)

	case !supported && unsupported:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			fmt.Sprintf("`aggregate` stage %q is not implemented yet", name),
			name+" (stage)", // to differentiate update operator $set from aggregation stage $set, etc
		)

	case !supported && !unsupported:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupInvalidAccumulator,
			fmt.Sprintf("Unrecognized pipeline stage name: %q", name),
			name+" (stage)", // to differentiate update operator $set from aggregation stage $set, etc
		)
	}

	panic("not reached")
}
