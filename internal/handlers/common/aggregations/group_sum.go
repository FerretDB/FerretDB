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

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// sumAccumulator represents $sum accumulator for $group.
type sumAccumulator struct {
	expression types.Expression
	count      any
}

// newSumAccumulator creates a new $sum accumulator for $group.
func newSumAccumulator(accumulation *types.Document) (Accumulator, error) {
	expr := must.NotFail(accumulation.Get("$sum"))
	accumulator := new(sumAccumulator)

	switch expr := expr.(type) {
	case *types.Array:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupUnaryOperator,
			"The $sum accumulator is a unary operator",
			"$sum (accumulator)",
		)
	case float64:
		accumulator.count = expr
	case string:
		var err error
		if accumulator.expression, err = types.NewExpression(expr); err != nil {
			// $sum returns 0 on non-existent field.
			accumulator.count = int32(0)
		}
	case int32, int64:
		accumulator.count = expr
	default:
		accumulator.count = int32(0)
		// $sum returns 0 on non-numeric field
	}

	return accumulator, nil
}

// Accumulate implements Accumulator interface.
func (s *sumAccumulator) Accumulate(ctx context.Context, groupID any, grouped []*types.Document) (any, error) {
	if s.expression != nil {
		var values []any

		for _, doc := range grouped {
			v := s.expression.Evaluate(doc)
			values = append(values, v)
		}

		res := sumNumbers(values...)

		return res, nil
	}

	switch count := s.count.(type) {
	case float64:
		return float64(len(grouped)) * count, nil
	case int32:
		return int32(len(grouped)) * count, nil
	case int64:
		return int64(len(grouped)) * count, nil
	}

	// $sum returns 0 on non-existent and non-numeric field.
	return int32(0), nil
}

// check interfaces
var (
	_ Accumulator = (*sumAccumulator)(nil)
)
