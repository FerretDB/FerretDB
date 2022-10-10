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
		if !utf8.ValidString(key) {
			return fmt.Errorf("invalid key: %q (not a valid UTF-8 string)", key)
		}

		if strings.Contains(key, "$") {
			return fmt.Errorf("invalid key: %q (key mustn't contain a string)", key)
		}
	}

	return nil
}

// checkMismatch checks if the document is logically correct.
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
