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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
)

// sort represents $sort stage.
type sort struct {
	fields *types.Document
}

// newSort creates a new $sort stage.
func newSort(stage *types.Document) (Stage, error) {
	fields, err := common.GetRequiredParam[*types.Document](stage, "$sort")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrSortBadExpression,
			"the $sort key specification must be an object",
			"$sort",
		)
	}

	if fields.Len() == 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrSortMissingKey,
			"$sort stage must have at least one sort key",
			"$sort",
		)
	}

	// TODO: https://github.com/FerretDB/FerretDB/issues/2090

	return &sort{
		fields: fields,
	}, nil
}

// Process implements Stage interface.
func (m *sort) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	if err := common.SortDocuments(in, m.fields); err != nil {
		return nil, err
	}

	return in, nil
}

// check interfaces
var (
	_ Stage = (*sort)(nil)
)
