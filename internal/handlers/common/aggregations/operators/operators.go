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

// Package operators provides aggregation operators.
package operators

import (
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
)

// newAccumulatorFunc is a type for a function that creates an accumulation operator.
type newAccumulatorFunc func(expression *types.Document) (Accumulator, error)

// Accumulator is a common interface for aggregation accumulation operators.
type Accumulator interface {
	// Accumulate documents and returns the result of applying operator.
	Accumulate(ctx context.Context, in []*types.Document) (any, error)
}

// GetAccumulator returns accumulator for provided value v with key.
// TODO consider better design
func GetAccumulator(stage, key string, v any) (Accumulator, error) {
	accumulation, ok := v.(*types.Document)
	if !ok || accumulation.Len() == 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupInvalidAccumulator,
			fmt.Sprintf("The field '%s' must be an accumulator object", key),
			stage+" (stage)",
		)
	}

	// accumulation document contains only one field.
	if accumulation.Len() > 1 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupMultipleAccumulator,
			fmt.Sprintf("The field '%s' must specify one accumulator", key),
			stage+" (stage)",
		)
	}

	operator := accumulation.Command()

	newAccumulator, ok := Accumulators[operator]
	if !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNotImplemented,
			fmt.Sprintf("%s accumulator %q is not implemented yet", stage, operator),
			operator+" (accumulator)",
		)
	}

	return newAccumulator(accumulation)
}

// Accumulators maps all aggregation accumulators.
var Accumulators = map[string]newAccumulatorFunc{
	// sorted alphabetically
	"$count": newCount,
	"$sum":   newSum,
	// please keep sorted alphabetically
}

// newOperatorFunc is a type for a function that creates a standard aggregation operator.
type newOperatorFunc func(expression *types.Document) (Operator, error)

// Operator is a common interface for standard aggregation operators.
// TODO consider not creating operators as structs - they don't require to store the state.
type Operator interface {
	// Process document and returns the result of applying operator.
	//
	// TODO make sure that operators work always the same for every stage,
	// if not - provide calling stage as an argument.
	Process(ctx context.Context, in *types.Document) (any, error)
}

// Operators maps all standard aggregation operators.
var Operators = map[string]newOperatorFunc{
	// sorted alphabetically
	// please keep sorted alphabetically
}
