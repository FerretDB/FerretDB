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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// newOperatorFunc is a type for a function that creates a standard aggregation operator.
type newOperatorFunc func(expression *types.Document) (Operator, error)

// Operator is a common interface for standard aggregation operators.
type Operator interface {
	// Process document and returns the result of applying operator.
	Process(in *types.Document) (any, error)
}

// Operators maps all standard aggregation operators.
var Operators = map[string]newOperatorFunc{
	// sorted alphabetically
	// please keep sorted alphabetically
}

// Get returns operator for provided value.
func Get(value any, stage, key string) (Operator, error) {
	operatorDoc, ok := value.(*types.Document)
	if !ok || operatorDoc.Len() == 0 {
		return nil, lazyerrors.New(
			fmt.Sprintf("The field '%s' must be an object", key),
		)
	}

	// operator document contains only one field.
	if operatorDoc.Len() > 1 {
		return nil, lazyerrors.New(
			fmt.Sprintf("The field '%s' must specify one operator", key),
		)
	}

	operator := operatorDoc.Command()

	newOperator, ok := Operators[operator]
	if !ok {
		return nil, lazyerrors.New(
			fmt.Sprintf("%s operator %q is not implemented yet", stage, operator),
		)
	}

	return newOperator(operatorDoc)
}
