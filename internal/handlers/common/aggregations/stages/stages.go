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
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// newStageFunc is a type for a function that creates a new aggregation stage.
type newStageFunc func(stage *types.Document) (Stage, error)

// StageType is a type for aggregation stage types.
type StageType int

const (
	// StageTypeDocuments is a type for stages that process documents.
	StageTypeDocuments StageType = iota

	// StageTypeStats is a type for stages that process statistics and doesn't need documents.
	StageTypeStats
)

// Stage is a common interface for all aggregation stages.
type Stage interface {
	// Process applies an aggregate stage on documents from iterator.
	Process(ctx context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error)

	// Type returns the type of the stage.
	//
	// TODO Remove it? https://github.com/FerretDB/FerretDB/issues/2423
	Type() StageType
}

// stages maps all supported aggregation stages.
var stages = map[string]newStageFunc{
	// sorted alphabetically
	"$collStats": newCollStats,
	"$count":     newCount,
	"$group":     newGroup,
	"$limit":     newLimit,
	"$match":     newMatch,
	"$skip":      newSkip,
	"$sort":      newSort,
	"$unwind":    newUnwind,
	// please keep sorted alphabetically
}

// unsupportedStages maps all unsupported yet stages.
var unsupportedStages = map[string]struct{}{
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
	"$limit":                  {},
	"$listLocalSessions":      {},
	"$listSessions":           {},
	"$lookup":                 {},
	"$merge":                  {},
	"$out":                    {},
	"$planCacheStats":         {},
	"$project":                {},
	"$redact":                 {},
	"$replaceRoot":            {},
	"$replaceWith":            {},
	"$sample":                 {},
	"$search":                 {},
	"$searchMeta":             {},
	"$set":                    {},
	"$setWindowFields":        {},
	"$sharedDataDistribution": {},
	"$skip":                   {},
	"$sortByCount":            {},
	"$unionWith":              {},
	"$unset":                  {},
}

// NewStage creates a new aggregation stage.
func NewStage(stage *types.Document) (Stage, error) {
	if stage.Len() != 1 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageInvalid,
			"A pipeline stage specification object must contain exactly one field.",
			"aggregate",
		)
	}

	name := stage.Command()

	f, ok := stages[name]
	if !ok {
		if _, ok := unsupportedStages[name]; ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				fmt.Sprintf("`aggregate` stage %q is not implemented yet", name),
				name+" (stage)", // to differentiate update operator $set from aggregation stage $set, etc
			)
		}

		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupInvalidAccumulator,
			fmt.Sprintf("Unrecognized pipeline stage name: %q", name),
			name+" (stage)", // to differentiate update operator $set from aggregation stage $set, etc
		)
	}

	return f(stage)
}
