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

type ValidationError struct {
	Reason error
}

func NewValidationError(reason error) error {
	return &ValidationError{Reason: reason}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Invalid document, reason: %s.", e.Reason)
}

func (e *ValidationError) Unwrap() error {
	return e.Reason
}

// Errors might be wrapped, so the caller needs to use errors.Is to check the error,
// for example, errors.Is(err, ErrInvalidKey).
var (
	// ErrInvalidKey indicates that a key didn't passed checks.
	ErrInvalidKey = fmt.Errorf("invalid key")
)

// ValidateData checks if the document represents a valid "data document".
func (d *Document) ValidateData() error {
	// The following bloc should be used to checks that keys are valid.
	// All further key related validation rules should be added here.
	for _, key := range d.keys {
		if !utf8.ValidString(key) {
			return fmt.Errorf("%w: %q (not a valid UTF-8 string)", ErrInvalidKey, key)
		}

		if strings.Contains(key, "$") {
			return fmt.Errorf("%w: %q (key mustn't contain $)", ErrInvalidKey, key)
		}
	}

	return nil
}
