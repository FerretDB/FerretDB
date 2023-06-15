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

// operatorErrorCode represents the type of error.
type operatorErrorCode uint

const (
	_ operatorErrorCode = iota

	// ErrArgsInvalidLen indicates that operator have invalid amount of arguments.
	ErrArgsInvalidLen

	// ErrWrongType indicates that operator field is not a document.
	ErrWrongType

	// ErrEmptyField indicates that operator field is empty.
	ErrEmptyField

	// ErrTooManyFields indicates that operator field specifes more than one operators.
	ErrTooManyFields

	// ErrNotImplemented indicates that given operator is not implemented yet.
	ErrNotImplemented

	// ErrInvalidExpression indicates that given operator does not exist.
	ErrInvalidExpression

	// ErrInvalidNestedExpression indicates that operator inside the target operator does not exist.
	ErrInvalidNestedExpression

	// ErrNoOperator indicates that given document does not contain any operator.
	ErrNoOperator
)

// newOperatorError returns new OperatorError.
func newOperatorError(code operatorErrorCode, msg string) error {
	return OperatorError{
		code: code,
		msg:  msg,
	}
}

// OperatorError is used for reporting operator errors.
type OperatorError struct {
	msg  string
	code operatorErrorCode
}

// Error implements error interface.
func (opErr OperatorError) Error() string {
	return opErr.msg
}

// Code returns operatorError code.
func (opErr OperatorError) Code() operatorErrorCode {
	return opErr.code
}
