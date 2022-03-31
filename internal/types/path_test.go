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

	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestRemoveByPath(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		path []string
		res  any
		err  string
	}
	for _, tc := range []testCase{{ //nolint:paralleltest // false positive
		name: "not found no error, ismaster field removed",
		path: []string{"ismaster", "0"},
		res: MustNewDocument(
			"ismaster", true,
			"client", MustNewArray(
				MustNewDocument(
					"document", "abc",
					"score", float64(42.13),
					"age", int32(1000),
				),
				MustNewDocument(
					"document", "def",
					"score", float64(42.13),
					"age", int32(1000),
				),
				MustNewDocument(
					"document", "jkl",
					"score", int32(24),
					"age", int32(1002),
				),
			),
			"value", must.NotFail(NewArray("none")),
		),
	}, {
		name: "removed entire client field",
		path: []string{"client"},
		res: MustNewDocument(
			"ismaster", true,
			"value", must.NotFail(NewArray("none")),
		),
	}, {
		name: "only 1d array element of client field is removed",
		path: []string{"client", "1"},
		res: MustNewDocument(
			"ismaster", true,
			"client", MustNewArray(
				MustNewDocument(
					"document", "abc",
					"score", float64(42.13),
					"age", int32(1000),
				),
				MustNewDocument(
					"document", "jkl",
					"score", int32(24),
					"age", int32(1002),
				),
			),
			"value", must.NotFail(NewArray("none")),
		),
	}, {
		name: "not found, element must be on place, no error",
		path: []string{"client", "3"},
		res: MustNewDocument(
			"ismaster", true,
			"client", MustNewArray(
				MustNewDocument(
					"document", "abc",
					"score", float64(42.13),
					"age", int32(1000),
				),
				MustNewDocument(
					"document", "def",
					"score", float64(42.13),
					"age", int32(1000),
				),
				MustNewDocument(
					"document", "jkl",
					"score", int32(24),
					"age", int32(1002),
				),
			),
			"value", must.NotFail(NewArray("none")),
		),
	}, {
		name: "not found, no error doc is same",
		path: []string{"compression", "invalid"},
		res: MustNewDocument(
			"ismaster", true,
			"client", MustNewArray(
				MustNewDocument(
					"document", "abc",
					"score", float64(42.13),
					"age", int32(1000),
				),
				MustNewDocument(
					"document", "def",
					"score", float64(42.13),
					"age", int32(1000),
				),
				MustNewDocument(
					"document", "jkl",
					"score", int32(24),
					"age", int32(1002),
				),
			),
			"value", must.NotFail(NewArray("none")),
		),
	}} {
		tc := tc
		t.Run(fmt.Sprint(tc.path), func(t *testing.T) {
			t.Parallel()

			doc := MustNewDocument(
				"ismaster", true,
				"client", MustNewArray(
					MustNewDocument(
						"document", "abc",
						"score", float64(42.13),
						"age", int32(1000),
					),
					MustNewDocument(
						"document", "def",
						"score", float64(42.13),
						"age", int32(1000),
					),
					MustNewDocument(
						"document", "jkl",
						"score", int32(24),
						"age", int32(1002),
					),
				),
				"value", must.NotFail(NewArray("none")),
			)

			doc.RemoveByPath(tc.path...)
			assert.Equal(t, tc.res, doc, tc.name)
		})
	}
}

func TestGetByPath(t *testing.T) {
	t.Parallel()

	doc := MustNewDocument(
		"ismaster", true,
		"client", MustNewDocument(
			"driver", MustNewDocument(
				"name", "nodejs",
				"version", "4.0.0-beta.6",
			),
			"os", MustNewDocument(
				"type", "Darwin",
				"name", "darwin",
				"architecture", "x64",
				"version", "20.6.0",
			),
			"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
			"application", MustNewDocument(
				"name", "mongosh 1.0.1",
			),
		),
		"compression", must.NotFail(NewArray("none")),
		"loadBalanced", false,
	)

	type testCase struct {
		path []string
		res  any
		err  string
	}

	for _, tc := range []testCase{{ //nolint:paralleltest // false positive
		path: []string{"compression", "0"},
		res:  "none",
	}, {
		path: []string{"compression"},
		res:  must.NotFail(NewArray("none")),
	}, {
		path: []string{"client", "driver"},
		res: MustNewDocument(
			"name", "nodejs",
			"version", "4.0.0-beta.6",
		),
	}, {
		path: []string{"client", "0"},
		err:  `types.getByPath: types.Document.Get: key not found: "0"`,
	}, {
		path: []string{"compression", "invalid"},
		err:  `types.getByPath: strconv.Atoi: parsing "invalid": invalid syntax`,
	}, {
		path: []string{"client", "missing"},
		err:  `types.getByPath: types.Document.Get: key not found: "missing"`,
	}, {
		path: []string{"compression", "1"},
		err:  `types.getByPath: types.Array.Get: index 1 is out of bounds [0-1)`,
	}, {
		path: []string{"compression", "0", "invalid"},
		err:  `types.getByPath: can't access string by path "invalid"`,
	}} {
		tc := tc
		t.Run(fmt.Sprint(tc.path), func(t *testing.T) {
			t.Parallel()

			res, err := getByPath(doc, tc.path...)
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
