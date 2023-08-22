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

// Package accumulators provides aggregation accumulator operators.
// Accumulators are different from other operators as they perform operations
// on a group of documents rather than a single document.
// They are used only in small subset of available stages (like `$group`).
//
// Accumulators that can be used outside of accumulation with different behaviour (like `$sum`),
// should be stored in both operators and accumulators packages.
package accumulators

import (
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// newAccumulatorFunc is a type for a function that creates an accumulation operator.
type newAccumulatorFunc func(args ...any) (Accumulator, error)

// Accumulator is a common interface for aggregation accumulation operators.
type Accumulator interface {
	// Accumulate documents and returns the result of applying operator.
	// It should always close iterator.
	Accumulate(iter types.DocumentsIterator) (any, error)
}

// NewAccumulator returns accumulator for provided value.
func NewAccumulator(stage, key string, value any) (Accumulator, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/2689
	accumulation, ok := value.(*types.Document)
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

	expression := must.NotFail(accumulation.Get(operator))

	return newAccumulator(accumulation)
}

// Accumulators maps all aggregation accumulators.
var Accumulators = map[string]newAccumulatorFunc{
	// sorted alphabetically
	"$count": newCount,
	"$sum":   newSum,
	// please keep sorted alphabetically
}
