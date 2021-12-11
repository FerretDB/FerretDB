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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		doc  Document
		err  error
	}{{
		name: "normal",
		doc: Document{
			keys: []string{"0"},
			m:    map[string]any{"0": "foo"},
		},
	}, {
		name: "empty",
		doc:  Document{},
	}, {
		name: "different keys",
		doc: Document{
			keys: []string{"0"},
			m:    map[string]any{"1": "foo"},
		},
		err: fmt.Errorf(`types.Document.validate: key not found: "0"`),
	}, {
		name: "duplicate keys",
		doc: Document{
			keys: []string{"0", "0"},
			m:    map[string]any{"0": "foo"},
		},
		err: fmt.Errorf("types.Document.validate: keys and values count mismatch: 1 != 2"),
	}, {
		name: "duplicate and different keys",
		doc: Document{
			keys: []string{"0", "0"},
			m:    map[string]any{"0": "foo", "1": "bar"},
		},
		err: fmt.Errorf(`types.Document.validate: duplicate key: "0"`),
	}} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.doc.validate()
			assert.Equal(t, tc.err, err)
		})
	}
}
