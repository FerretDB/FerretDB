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

package types

import (
	"math"
	"math/big"
)

//go:generate ../../bin/stringer -linecomment -type NumberErrorCode

// NumberErrorCode represents error code from numerical operation.
type NumberErrorCode int

const (
	_ NumberErrorCode = iota

	// ErrLongExceeded indicates that long exceeded in its size.
	ErrLongExceeded

	// ErrNotExactResult indicates that float addition dropped precision.
	ErrNotExactResult
)

// NumberError describes an error that occurs applying number operation.
type NumberError struct {
	code NumberErrorCode
}

// newNumberError creates a new NumberError.
func newNumberError(code NumberErrorCode) error {
	return &NumberError{code: code}
}

// Error implements the error interface.
func (e *NumberError) Error() string {
	return e.code.String()
}

// Code returns the FieldPathError code.
func (e *NumberError) Code() NumberErrorCode {
	return e.code
}

// AddNumbers returns the result of addition and error if addition failed.
func AddNumbers(vs ...any) (any, error) {
	sum := big.NewInt(0)
	sumFloat := big.NewFloat(0)

	var hasFloat64, hasInt64 bool

	for _, v := range vs {
		switch v := v.(type) {
		case float64:
			hasFloat64 = true
			if v > MaxSafeDouble {
				// todo handle lost precision
				smallPart := v - MaxSafeDouble
				sum.Add(sum, big.NewInt(int64(MaxSafeDouble)))
				sumFloat.Add(sumFloat, big.NewFloat(smallPart))
				continue
			}

			if v < -MaxSafeDouble {
				// todo handle lost precision
				smallPart := v + MaxSafeDouble

				sum.Add(sum, big.NewInt(int64(-MaxSafeDouble)))
				sumFloat.Add(sumFloat, big.NewFloat(smallPart))
				continue
			}

			// todo check overflow
			sumFloat.Add(sumFloat, big.NewFloat(v))
		case int32:
			sum.Add(sum, big.NewInt(int64(v)))
		case int64:
			hasInt64 = true
			sum.Add(sum, big.NewInt(v))
		default:
			// ignore non-number
		}
	}

	if !sum.IsInt64() {
		return nil, newNumberError(ErrLongExceeded)
	}

	sumBig := sum.Int64()

	res := sumBig

	if hasFloat64 {
		f, accuracy := sumFloat.Float64()
		if accuracy != big.Exact {
			return nil, newNumberError(ErrNotExactResult)
		}

		// todo check overflow
		if sumBig > int64(MaxSafeDouble) || sumBig < -int64(MaxSafeDouble) {
			// not accurate result
			return float64(sumBig) + f, nil
		}

		return float64(sumBig) + f, nil
	}

	if hasInt64 {
		return res, nil
	}

	if res < math.MaxInt32 && res > math.MinInt32 {
		return int32(res), nil
	}

	return res, nil
}
