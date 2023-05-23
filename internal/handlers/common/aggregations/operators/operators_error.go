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
// Operators are used in aggregation stages to filter and model data.
// This package contains all operators apart from the accumulation operators,
// which are stored and described in accumulators package.
//
// Accumulators that can be used outside of accumulation with different behaviour (like `$sum`),
// should be stored in both operators and accumulators packages.
package operators

import "fmt"

// OperatorErrorCode represents aggregation operator error code.
type OperatorErrorCode int

const (
	// ErrWrongType indicates that operator field is not a document.
	ErrWrongType OperatorErrorCode = iota + 1

	// ErrEmptyField indicates that operator field does not specify any operator.
	ErrEmptyField

	// ErrTooManyFields indicates that operator field specifes more than one operators.
	ErrTooManyFields

	// ErrNotImplemented indicates that given operator is not implemented yet.
	ErrNotImplemented
)

// newExpressionError creates a new ExpressionError.
func newOperatorError(code OperatorErrorCode) error {
	return &OperatorError{code: code}
}

// OperatorError describes an error that occurs while evaluating operator.
type OperatorError struct {
	code OperatorErrorCode
}

// Error implements the error interface.
func (e *OperatorError) Error() string {
	return fmt.Sprint(e.code)
}

// Code returns the OperatorError code.
func (e *OperatorError) Code() OperatorErrorCode {
	return e.code
}
