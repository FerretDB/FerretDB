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

	"github.com/FerretDB/FerretDB/internal/util/must"
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
// If the document is not valid it returns *ValidationError or generic error message.
func (d *Document) ValidateData() error {
	keys := d.Keys()

	duplicateChecker := make(map[string]struct{}, len(keys))
	var idPresent bool

	// TODO: make sure that `_id` is the first item in the map
	for _, key := range keys {
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

		if _, ok := duplicateChecker[key]; ok {
			return newValidationError(fmt.Errorf("invalid key: %q (duplicate keys are not allowed)", key))
		}
		duplicateChecker[key] = struct{}{}

		if key == "_id" {
			idPresent = true
		}

		value := must.NotFail(d.Get(key))

		// TODO Add dance tests for infinity: https://github.com/FerretDB/FerretDB/issues/1151
		if v, ok := value.(float64); ok && math.IsInf(v, 0) {
			return newValidationError(fmt.Errorf("invalid value: { %q: %f } (infinity values are not allowed)", key, v))
		}
	}

	if !idPresent {
		return newValidationError(fmt.Errorf("invalid document: document must contain '_id' field"))
	}

	v := must.NotFail(d.Get("_id"))

	switch v.(type) {
	case *Array:
		return fmt.Errorf("The '_id' value cannot be of type array")
	case Regex:
		return fmt.Errorf("The '_id' value cannot be of type regex")
	}

	return nil
}
