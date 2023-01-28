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

package aggregations

import (
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
)

type newStage func(stage *types.Document) (Stage, error)

// Stage is a common interface for all aggregation stages.
//
// TODO use iterators instead of slices of documents.
type Stage interface {
	Process(ctx context.Context, in []*types.Document) ([]*types.Document, error)
}

var stages = map[string]newStage{
	// please keep sorted
	"$count": newCount,
	"$match": newMatch,
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
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			fmt.Sprintf("`aggregate` stage %q is not implemented yet", name),
			name,
		)
	}

	return f(stage)
}
