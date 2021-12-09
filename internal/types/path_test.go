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
	"github.com/stretchr/testify/require"
)

func TestGetByPath(t *testing.T) {
	t.Parallel()

	doc := MustMakeDocument(
		"ismaster", true,
		"client", MustMakeDocument(
			"driver", MustMakeDocument(
				"name", "nodejs",
				"version", "4.0.0-beta.6",
			),
			"os", MustMakeDocument(
				"type", "Darwin",
				"name", "darwin",
				"architecture", "x64",
				"version", "20.6.0",
			),
			"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
			"application", MustMakeDocument(
				"name", "mongosh 1.0.1",
			),
		),
		"compression", Array{"none"},
		"loadBalanced", false,
	)

	type testCase struct {
		path []any
		res  any
		err  string
	}

	for _, tc := range []testCase{{
		path: []any{"compression", 0},
		res:  "none",
	}, {
		path: []any{"compression"},
		res:  Array{"none"},
	}, {
		path: []any{"client", "driver"},
		res: MustMakeDocument(
			"name", "nodejs",
			"version", "4.0.0-beta.6",
		),
	}, {
		path: []any{"client", 0},
		err:  `types.GetByPath: can't access types.Document by path 0 (int)`,
	}, {
		path: []any{"compression", "invalid"},
		err:  `types.GetByPath: can't access types.Array by path invalid (string)`,
	}, {
		path: []any{"client", "missing"},
		err:  `types.GetByPath: types.Document.Get: key not found: "missing"`,
	}, {
		path: []any{"compression", 1},
		err:  `types.GetByPath: types.Array.Get: index 1 is out of bounds [0-1)`,
	}, {
		path: []any{"compression", 0, "invalid"},
		err:  `types.GetByPath: can't access string by path invalid (string)`,
	}} {
		tc := tc
		t.Run(fmt.Sprint(tc.path), func(t *testing.T) {
			res, err := GetByPath(doc, tc.path...)
			if tc.err == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.res, res)
			} else {
				assert.Empty(t, res)
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}
