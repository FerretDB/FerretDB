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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// limit represents $limit stage.
type limit struct {
	max int64
}

// newLimit creates a new $limit stage.
func newLimit(stage *types.Document) (Stage, error) {
	doc, err := stage.Get("$limit")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	max, err := common.GetWholeNumberParam(doc)
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrLimitStageInvalidArg,
			fmt.Sprintf("Invalid argument to $limit stage: Expected a number in: $limit: %v", doc),
		)
	}

	return &limit{
		max: max,
	}, nil
}

// Process implements Stage interface.
func (l *limit) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	return common.LimitDocuments(in, l.max)
}
