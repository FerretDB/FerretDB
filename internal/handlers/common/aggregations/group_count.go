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

// countAccumulator represents $count accumulator for $group.
type countAccumulator struct{}

// newCountAccumulator creates a new $count accumulator for $group.
func newCountAccumulator(accumulation *types.Document) (Accumulator, error) {
	expression, err := common.GetRequiredParam[*types.Document](accumulation, "$count")
	if err != nil || expression.Len() != 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			"$count takes no arguments, i.e. $count:{}",
			"$count (accumulator)",
		)
	}

	return new(countAccumulator), nil
}

// Accumulate implements Accumulator interface.
func (c *countAccumulator) Accumulate(ctx context.Context, grouped []*types.Document) (any, error) {
	return int32(len(grouped)), nil
}

// check interfaces
var (
	_ Accumulator = (*countAccumulator)(nil)
)
