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

package handlererrors

// ValidationError is used for reporting validation errors.
type ValidationError struct {
	err error
}

// Error implements [error].
func (v *ValidationError) Error() string {
	return v.err.Error()
}

// NewValidationError returns new ValidationError.
//
// Remove and make callers use validateValue only?
// TODO https://github.com/FerretDB/FerretDB/issues/2412
func NewValidationError(err error) error {
	return &ValidationError{err: err}
}

// check interfaces
var (
	_ error = (*ValidationError)(nil)
)
