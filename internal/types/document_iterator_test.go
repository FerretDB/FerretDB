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
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestDocumentIterator(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		doc      *Document
		expected []field
	}{
		"empty": {
			doc:      must.NotFail(NewDocument()),
			expected: []field{},
		},
		"one": {
			doc:      must.NotFail(NewDocument("foo", "bar")),
			expected: []field{{key: "foo", value: "bar"}},
		},
		"two": {
			doc:      must.NotFail(NewDocument("foo", "bar", "baz", "qux")),
			expected: []field{{key: "foo", value: "bar"}, {key: "baz", value: "qux"}},
		},
		"duplicates": {
			doc:      must.NotFail(NewDocument("foo", "bar", "baz", "qux", "foo", "quuz")),
			expected: []field{{key: "foo", value: "bar"}, {key: "baz", value: "qux"}, {key: "foo", value: "quuz"}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			iter := tc.doc.Iterator()
			defer iter.Close()

			for i := 0; i < len(tc.expected); i++ {
				key, value, err := iter.Next()
				require.NoError(t, err)

				require.Equal(t, tc.expected[i].key, key)
				require.Equal(t, tc.expected[i].value, value)
			}

			_, _, err := iter.Next()
			assert.Equal(t, iterator.ErrIteratorDone, err)

			// check that Next() can be called again
			_, _, err = iter.Next()
			assert.Equal(t, iterator.ErrIteratorDone, err)
		})
	}
}
