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

// ValidateCmd checks if the document represents a valid command.
func (d *Document) ValidateCmd() error {
	if err := d.checkMismatch(); err != nil {
		panic(fmt.Sprintf("types.Document.ValidateCmd: %s", err))
	}

	return nil
}

// ValidateData checks if the document represents a valid "data document".
func (d *Document) ValidateData() error {
	if err := d.checkMismatch(); err != nil {
		panic(fmt.Sprintf("types.Document.ValidateData: %s", err))
	}

	// The following block checks that keys are valid.
	// All further key related validation rules should be added here.
	for _, key := range d.keys {
		if strings.Contains(key, "$") {
			return fmt.Errorf("key %q contains $", key)
		}
	}

	return nil
}

// checkMismatch checks if the document is "reasonable".
// It returns error if there is a potential bug in code: the doc is a nil pointer
// or the number of keys and values don't match.
func (d *Document) checkMismatch() error {
	if d == nil {
		return fmt.Errorf("d is nil")
	}

	if len(d.m) != len(d.keys) {
		return fmt.Errorf("keys and values count mismatch: %d != %d", len(d.m), len(d.keys))
	}

	return nil
}

// validate checks if the document is valid.
// todo: decide what to do with this function.
func (d *Document) validate() error {
	if d == nil {
		panic("types.Document.validate: d is nil")
	}

	if len(d.m) != len(d.keys) {
		return fmt.Errorf("types.Document.validate: keys and values count mismatch: %d != %d", len(d.m), len(d.keys))
	}

	// TODO check that _id is not regex or array: https://github.com/FerretDB/FerretDB/issues/1235

	prevKeys := make(map[string]struct{}, len(d.keys))
	for _, key := range d.keys {
		if !isValidKey(key) {
			return fmt.Errorf("types.Document.validate: invalid key: %q", key)
		}

		value, ok := d.m[key]
		if !ok {
			return fmt.Errorf("types.Document.validate: key not found: %q", key)
		}

		if _, ok := prevKeys[key]; ok {
			return fmt.Errorf("types.Document.validate: duplicate key: %q", key)
		}
		prevKeys[key] = struct{}{}

		if err := validateValue(value); err != nil {
			return fmt.Errorf("types.Document.validate: %w", err)
		}
	}

	return nil
}

// isValidKey returns false if key is not a valid document field key.
//
// TODO That function should be removed once we have separate validation for command and data documents.
func isValidKey(key string) bool {
	if key == "" {
		// TODO that should be valid only for command documents, not for data documents
		return true
	}

	// forbid keys like $k (used by fjson representation), but allow $db (used by many commands)
	if key[0] == '$' && len(key) <= 2 {
		return false
	}

	return utf8.ValidString(key)
}
