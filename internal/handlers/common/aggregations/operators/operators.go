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
