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

// newOperatorFunc is a type for a function that creates an operator.
type newOperatorFunc func(expression *types.Document) (Operator, error)

// Operator is a common interface for aggregation operators.
type Operator interface {
	// Accumulate documents and returns the result of applying accumulation operator.
	Accumulate(ctx context.Context, groupID any, in []*types.Document) (any, error)
}

// GroupOperators maps supported operators of $group aggregation stage.
var GroupOperators = map[string]newOperatorFunc{
	// sorted alphabetically
	"$count": newCount,
	"$sum":   newSum,
	// please keep sorted alphabetically
}
