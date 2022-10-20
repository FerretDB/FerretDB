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
	// The following block should be used to check that keys are valid.
	// All further key related validation rules should be added here.
	for _, key := range d.keys {
		// Tests for this case are in `dance`.
		if !utf8.ValidString(key) {
			return newValidationError(fmt.Errorf("invalid key: %q (not a valid UTF-8 string)", key))
		}

		// Tests for this case are in `dance`.
		if strings.Contains(key, "$") {
			return newValidationError(fmt.Errorf("invalid key: %q (key must not contain $)", key))
		}

		if _, ok := d.m[key]; ok && len(d.m[key]) > 1 {
			return newValidationError(fmt.Errorf("invalid key: %q (duplicate keys are not allowed)", key))
		}
	}

	return nil
}
