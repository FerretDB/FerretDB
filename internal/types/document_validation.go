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
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../bin/stringer -linecomment -type ValidationErrorCode

// ValidationErrorCode represents ValidationData error code.
type ValidationErrorCode int

const (
	_ ValidationErrorCode = iota

	// ErrValidation indicates that document is invalid.
	ErrValidation

	// ErrWrongIDType indicates that _id field is invalid.
	ErrWrongIDType

	// ErrIDNotFound indicates that _id field is not found.
	ErrIDNotFound
)

// ValidationError describes an error that could occur when validating a document.
type ValidationError struct {
	code   ValidationErrorCode
	reason error
}

// newValidationError creates a new ValidationError.
func newValidationError(code ValidationErrorCode, reason error) error {
	return &ValidationError{reason: reason, code: code}
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return e.reason.Error()
}

// Code returns the ValidationError code.
func (e *ValidationError) Code() ValidationErrorCode {
	return e.code
}

// ValidateData checks if the document represents a valid "data document".
// It places `_id` field into the fields slice 0 index.
// It replaces negative zero -0 with valid positive zero 0.
// If the document is not valid it returns *ValidationError.
func (d *Document) ValidateData() error {
	d.moveIDToTheFirstIndex()

	keys := d.Keys()
	values := d.Values()

	duplicateChecker := make(map[string]struct{}, len(keys))
	var idPresent bool

	for i, key := range keys {
		// integration tests for those cases are in the `dance` repo

		if !utf8.ValidString(key) {
			return newValidationError(ErrValidation, fmt.Errorf("invalid key: %q (not a valid UTF-8 string)", key))
		}

		if strings.HasPrefix(key, "$") {
			return newValidationError(ErrValidation, fmt.Errorf("invalid key: %q (key must not start with '$' sign)", key))
		}

		if strings.Contains(key, ".") {
			return newValidationError(ErrValidation, fmt.Errorf("invalid key: %q (key must not contain '.' sign)", key))
		}

		if _, ok := duplicateChecker[key]; ok {
			return newValidationError(ErrValidation, fmt.Errorf("invalid key: %q (duplicate keys are not allowed)", key))
		}
		duplicateChecker[key] = struct{}{}

		if key == "_id" {
			idPresent = true
		}

		value := values[i]

		switch v := value.(type) {
		case *Document:
			err := v.ValidateData()
			if err != nil {
				var vErr *ValidationError

				if errors.As(err, &vErr) && vErr.code == ErrIDNotFound {
					continue
				}

				return err
			}
		case *Array:
			if key == "_id" {
				return newValidationError(ErrWrongIDType, fmt.Errorf("The '_id' value cannot be of type array"))
			}

			for i := 0; i < v.Len(); i++ {
				item := must.NotFail(v.Get(i))

				switch item := item.(type) {
				case *Document:
					err := item.ValidateData()
					if err != nil {
						var vErr *ValidationError

						if errors.As(err, &vErr) && vErr.code == ErrIDNotFound {
							continue
						}

						return err
					}
				case *Array:
					return newValidationError(ErrValidation, fmt.Errorf(
						"invalid value: { %q: %v } (nested arrays are not supported)", key, FormatAnyValue(v),
					))
				case float64:
					if item == 0 && math.Signbit(item) {
						v.Set(i, math.Copysign(0, +1))
					}
				}
			}
		case float64:
			if math.IsInf(v, 0) {
				return newValidationError(
					ErrValidation, fmt.Errorf("invalid value: { %q: %f } (infinity values are not allowed)", key, v),
				)
			}

			if v == 0 && math.Signbit(v) {
				d.Set(key, math.Copysign(0, +1))
			}
		case Regex:
			if key == "_id" {
				return newValidationError(ErrWrongIDType, fmt.Errorf("The '_id' value cannot be of type regex"))
			}
		}
	}

	if !idPresent {
		return newValidationError(ErrIDNotFound, fmt.Errorf("invalid document: document must contain '_id' field"))
	}

	return nil
}
