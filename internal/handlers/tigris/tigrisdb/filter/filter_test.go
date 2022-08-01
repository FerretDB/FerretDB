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

package filter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
)

func TestExprBuild(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		expr     Expr
		expected json.RawMessage
	}{
		"string": {
			expr:     Eq("_id", `foo`),
			expected: json.RawMessage{},
		},
		"objectID": {
			expr:     Eq("_id", types.ObjectID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c}),
			expected: json.RawMessage{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := tc.expr.Build()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
