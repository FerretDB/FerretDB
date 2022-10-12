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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// TestDocumentValidateData covers ValidateData method.
// Proper testing of validation requires integration tests,
// see https://github.com/FerretDB/dance/tree/main/tests/diff for more examples.
func TestDocumentValidateData(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		doc     *Document
		isValid bool
	}{
		"Valid": {
			doc:     must.NotFail(NewDocument("foo", "bar")),
			isValid: true,
		},
		"KeyIsNotUTF8": {
			doc:     must.NotFail(NewDocument("\xF4\x90\x80\x80", "bar")), //  the key is out of range for UTF-8
			isValid: false,
		},
		"KeyContains$": {
			doc:     must.NotFail(NewDocument("$v", "bar")),
			isValid: false,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.doc.ValidateData()
			if tc.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
