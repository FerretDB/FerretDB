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
	return opErr.Error()
}
