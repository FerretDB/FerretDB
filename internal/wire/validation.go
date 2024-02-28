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

package wire

import (
	"errors"
	"math"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ValidationError is used for reporting validation errors.
type ValidationError struct {
	err error
}

// Error implements error interface.
func (v *ValidationError) Error() string {
	return v.err.Error()
}

// newValidationError returns new ValidationError.
//
// Remove and make callers use validateValue only?
// TODO https://github.com/FerretDB/FerretDB/issues/2412
func newValidationError(err error) error {
	return &ValidationError{err: err}
}

// validateValue checks given value and returns error if not supported value was encountered.
func validateValue(v any) error {
	switch v := v.(type) {
	case *types.Document:
		for _, v := range v.Values() {
			if err := validateValue(v); err != nil {
				return err
			}
		}

	case *types.Array:
		for i := 0; i < v.Len(); i++ {
			v := must.NotFail(v.Get(i))
			if err := validateValue(v); err != nil {
				return err
			}
		}

	case float64:
		if math.IsNaN(v) {
			return errors.New("NaN is not supported")
		}
	}

	return nil
}

// check interfaces
var (
	_ error = (*ValidationError)(nil)
)
