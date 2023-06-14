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

type OperatorErrorCode uint

const (
	_ OperatorErrorCode = iota
	ErrArgsInvalidLen

	// ErrWrongType indicates that operator field is not a document.
	ErrWrongType // Invalid type of operator field (expected document)

	// ErrEmptyField indicates that operator field does not specify any operator.
	ErrEmptyField // The operator field is empty (expected document)

	// ErrTooManyFields indicates that operator field specifes more than one operators.
	ErrTooManyFields // The operator field specifies more than one operator

	// ErrNotImplemented indicates that given operator is not implemented yet.
	ErrNotImplemented // The operator is not implemented yet

	// // ErrNotImplemented indicates that given operator does not exist.
	// ErrInvalidExpression // Unrecognized expression

	// ErrNoOperator indicates that given document does not contain any operator.
	ErrNoOperator // No operator in document
)

func NewOperatorError(code OperatorErrorCode, err error) error {
	if err == nil {
		panic("Provided err mustn't be nil")
	}

	return OperatorError{
		code: code,
		err:  err,
	}
}

type OperatorError struct {
	code OperatorErrorCode
	err  error
}

func (opErr OperatorError) Error() string {
	return opErr.err.Error()
}

func (opErr OperatorError) Code() OperatorErrorCode {
	return opErr.code
}
