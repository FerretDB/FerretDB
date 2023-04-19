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

	"github.com/FerretDB/FerretDB/internal/types"
)

// newAccumulatorFunc is a type for a function that creates an accumulation operator.
type newAccumulatorFunc func(expression *types.Document) (Accumulator, error)

// Accumulator is a common interface for aggregation accumulation operators.
type Accumulator interface {
	// Accumulate documents and returns the result of applying operator.
	Accumulate(ctx context.Context, groupID any, in []*types.Document) (any, error)
}

// GroupAccumulators maps supported accumulators of $group aggregation stage.
var GroupAccumulators = map[string]newAccumulatorFunc{
	// sorted alphabetically
	"$count": newCount,
	"$sum":   newSum,
	// please keep sorted alphabetically
}
