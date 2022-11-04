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
	"math"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ValidationError is used for reporting validation errors.
type ValidationError struct {
	msg string
}

// Error implements error interface.
func (v ValidationError) Error() string {
	return v.msg
}

// NewValidationError returns new ValidationError.
func NewValidationError(err error) error {
	return &ValidationError{msg: err.Error()}
}

// validateValue checks given value and return error if not supported value was encountered.
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
			if err := validateValue(must.NotFail(v.Get(i))); err != nil {
				return err
			}
		}
	case float64:
		if math.IsNaN(v) {
			return lazyerrors.Errorf("NaN is not supported")
		}

		if v == 0 && math.Signbit(0) {
			return lazyerrors.Errorf("-0 is not supported")
		}
	}

	return nil
}
