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

package validators

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/FerretDB/FerretDB/internal/types"
)

type Data struct{}

func (d *Data) Validate(document *types.Document) error {
	// The following block checks that keys are valid.
	// All further key related validation rules should be added here.
	for _, key := range document.Keys() {
		if err := d.ValidateKey(key); err != nil {
			return err
		}
	}

	return nil
}

func (d *Data) ValidateKey(key string) error {
	if !utf8.ValidString(key) {
		return fmt.Errorf("invalid key: %q (not a valid UTF-8 string)", key)
	}

	if strings.Contains(key, "$") {
		return fmt.Errorf("invalid key: %q (the key mustn't contain a string)", key)
	}

	return nil
}

func (d *Data) ValidateValue() error {
	return nil
}
