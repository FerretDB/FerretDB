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
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// ValidationError describes an error that could occur when validating a document.
type ValidationError struct {
	reason error
}

// newValidationError creates a new ValidationError.
func newValidationError(reason error) error {
	return &ValidationError{reason: reason}
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("Invalid document, reason: %s.", e.reason)
}

// ValidateData checks if the document represents a valid "data document".
// If the document is not valid it returns *ValidationError.
func (d *Document) ValidateData() error {
	var idPresent bool

	// TODO: make sure that `_id` is the first item in the map
	//
	// The following block should be used to checks that keys and values are valid.
	// All further validation rules should be added here.
	for key, v := range d.m {
		// Tests for this case are in `dance`.
		if !utf8.ValidString(key) {
			return newValidationError(fmt.Errorf("invalid key: %q (not a valid UTF-8 string)", key))
		}

		// Tests for this case are in `dance`.
		if strings.Contains(key, "$") {
			return newValidationError(fmt.Errorf("invalid key: %q (key must not contain '$' sign)", key))
		}

		// Tests for this case are in `dance`.
		if strings.Contains(key, ".") {
			return newValidationError(fmt.Errorf("invalid key: %q (key must not contain '.' sign)", key))
		}

		if v, ok := v.(float64); ok && math.IsInf(v, 0) {
			return newValidationError(fmt.Errorf("invalid value: { %q: %f } (infinity values are not allowed)", key, v))
		}

		if key == "_id" {
			idPresent = true
		}
	}

	if !idPresent {
		return newValidationError(fmt.Errorf("invalid document: document must contain '_id' field"))
	}

	return nil
}
