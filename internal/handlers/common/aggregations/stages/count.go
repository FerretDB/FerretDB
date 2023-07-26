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
	"context"
	"strings"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// count represents $count stage.
type count struct {
	field string
}

// newCount creates a new $count stage.
func newCount(stage *types.Document) (aggregations.Stage, error) {
	field, err := common.GetRequiredParam[string](stage, "$count")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageCountNonString,
			"the count field must be a non-empty string",
			"$count (stage)",
		)
	}

	if len(field) == 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageCountNonEmptyString,
			"the count field must be a non-empty string",
			"$count (stage)",
		)
	}

	if strings.Contains(field, ".") {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageCountBadValue,
			"the count field cannot contain '.'",
			"$count (stage)",
		)
	}

	if strings.HasPrefix(field, "$") {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageCountBadPrefix,
			"the count field cannot be a $-prefixed path",
			"$count (stage)",
		)
	}

	if field == "_id" {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupID,
			"a group's _id may only be specified once",
			"$count (stage)",
		)
	}

	return &count{
		field: field,
	}, nil
}

// Process implements Stage interface.
func (c *count) Process(ctx context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error) { //nolint:lll // for readability
	return common.CountIterator(iter, closer, c.field), nil
}

// check interfaces
var (
	_ aggregations.Stage = (*count)(nil)
)
