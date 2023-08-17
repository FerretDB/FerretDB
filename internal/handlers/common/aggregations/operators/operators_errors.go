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

	// ErrTooManyFields indicates that operator field specifes more than one operators.
	ErrTooManyFields

	// ErrNotImplemented indicates that given operator is not implemented yet.
	ErrNotImplemented

	// ErrInvalidExpression indicates that given operator does not exist.
	ErrInvalidExpression

	// ErrInvalidNestedExpression indicates that operator inside the target operator does not exist.
	ErrInvalidNestedExpression
)

// newOperatorError returns new OperatorError.
func newOperatorError(code operatorErrorCode, name, msg string) error {
	return OperatorError{
		code: code,
		name: name,
		msg:  msg,
	}
}

// OperatorError is used for reporting operator errors.
type OperatorError struct {
	msg  string
	name string
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

// Name returns the name of the operator (e.g. $sum) that produced an error.
func (opErr OperatorError) Name() string {
	return opErr.name
}
