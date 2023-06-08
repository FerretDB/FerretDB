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

const (
	_ OperatorErrorCode = iota

	// ErrWrongType indicates that operator field is not a document.
	ErrWrongType // Invalid type of operator field (expected document)

	// ErrEmptyField indicates that operator field does not specify any operator.
	ErrEmptyField // The operator field is empty (expected document)

	// ErrTooManyFields indicates that operator field specifes more than one operators.
	ErrTooManyFields // The operator field specifies more than one operator

	// ErrNotImplemented indicates that given operator is not implemented yet.
	ErrNotImplemented // The operator is not implemented yet

	// ErrNotImplemented indicates that given operator does not exist.
	ErrInvalidExpression // Unrecognized expression

	// ErrNoOperator indicates that given document does not contain any operator.
	ErrNoOperator // No operator in document
)

type OperatorErrorCode uint

func NewOperatorError(code OperatorErrorCode, operator string, err error) OperatorError {
	return OperatorError{code: code, reason: err, operator: operator}
}

type OperatorError struct {
	code     OperatorErrorCode
	reason   error
	operator string
}

func (err OperatorError) Error() string {
	if err.operator == "" {
		return fmt.Sprintf("%s", err.reason)
	}

	return fmt.Sprintf("%s: %s", err.operator, err.reason)
}

func (err OperatorError) Code() OperatorErrorCode {
	return err.code
}

func (err OperatorError) Operator() string {
	return err.operator
}
