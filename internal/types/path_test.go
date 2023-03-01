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

	deepDoc := must.NotFail(NewDocument(
		"xxx", "yyy",
		"bar", float64(42.13),
		"baz", must.NotFail(NewDocument(
			"foo", "bar",
			"bar", float64(42.13),
			"baz", must.NotFail(NewDocument(
				"foo", "baz",
				"bar", float64(42.13),
				"baz", int32(1000),
			)),
		)),
	))

	sourceDoc := must.NotFail(NewDocument(
		"ismaster", true,
		"client", must.NotFail(NewArray(
			must.NotFail(NewDocument(
				"document", "abc",
				"score", float64(42.13),
				"age", int32(1000),
				"foo", deepDoc.DeepCopy(),
			)),
			must.NotFail(NewDocument(
				"document", "def",
				"score", float64(42.13),
				"age", int32(1000),
			)),
			must.NotFail(NewDocument(
				"document", "jkl",
				"score", int32(24),
				"age", int32(1002),
			)),
		)),
		"value", must.NotFail(NewArray("none")),
	))

	//nolint:paralleltest // false positive
	for name, tc := range map[string]struct {
		path Path
		res  *Document
	}{
		"test deep removal ok": {
			path: NewStaticPath("client", "0", "foo", "baz", "baz", "baz"),
			res: must.NotFail(NewDocument(
				"ismaster", true,
				"client", must.NotFail(NewArray(
					must.NotFail(NewDocument(
						"document", "abc",
						"score", float64(42.13),
						"age", int32(1000),
						"foo", must.NotFail(NewDocument(
							"xxx", "yyy",
							"bar", float64(42.13),
							"baz", must.NotFail(NewDocument(
								"foo", "bar",
								"bar", float64(42.13),
								"baz", must.NotFail(NewDocument(
									"foo", "baz",
									"bar", float64(42.13),
								)),
							)),
						)),
					)),
					must.NotFail(NewDocument(
						"document", "def",
						"score", float64(42.13),
						"age", int32(1000),
					)),
					must.NotFail(NewDocument(
						"document", "jkl",
						"score", int32(24),
						"age", int32(1002),
					)),
				)),
				"value", must.NotFail(NewArray("none")),
			)),
		},
		"not found no error": {
			path: NewStaticPath("ismaster", "0"),
			res:  sourceDoc.DeepCopy(),
		},
		"removed entire client field": {
			path: NewStaticPath("client"),
			res: must.NotFail(NewDocument(
				"ismaster", true,
				"value", must.NotFail(NewArray("none")),
			)),
		},
		"only 1d array element of client field is removed": {
			path: NewStaticPath("client", "1"),
			res: must.NotFail(NewDocument(
				"ismaster", true,
				"client", must.NotFail(NewArray(
					must.NotFail(NewDocument(
						"document", "abc",
						"score", float64(42.13),
						"age", int32(1000),
						"foo", deepDoc.DeepCopy(),
					)),
					must.NotFail(NewDocument(
						"document", "jkl",
						"score", int32(24),
						"age", int32(1002),
					)),
				)),
				"value", must.NotFail(NewArray("none")),
			)),
		},
		"not found, no error doc is same": {
			path: NewStaticPath("compression", "invalid"),
			res:  sourceDoc.DeepCopy(),
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doc := sourceDoc.DeepCopy()
			RemoveByPath(doc, tc.path)
			assert.Equal(t, tc.res, doc)
		})
	}
}

func TestRemoveByPathArray(t *testing.T) {
	t.Parallel()

	deepDoc := must.NotFail(NewDocument(
		"xxx", "yyy",
		"bar", float64(42.13),
		"baz", must.NotFail(NewDocument(
			"foo", "bar",
			"bar", float64(42.13),
			"baz", must.NotFail(NewDocument(
				"foo", "baz",
				"bar", float64(42.13),
				"baz", int32(1000),
			)),
		)),
	))

	src := must.NotFail(NewArray(
		"0", float64(42.13), int32(1000), "2",
		must.NotFail(NewDocument(
			"document", "abc",
			"score", float64(42.13),
			"age", int32(1000),
			"foo", deepDoc.DeepCopy(),
		)),
		must.NotFail(NewArray("1", "2", "3")),
	))

	for name, tc := range map[string]struct {
		path     Path
		expected *Array
	}{
		"array: remove by path": {
			path:     NewStaticPath("4"),
			expected: must.NotFail(NewArray("0", float64(42.13), int32(1000), "2", must.NotFail(NewArray("1", "2", "3")))),
		},
		"array: index exceeded": {
			path:     NewStaticPath("11"),
			expected: src.DeepCopy(),
		},
		"array: index is not number": {
			path:     NewStaticPath("abcd"),
			expected: src.DeepCopy(),
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			arr := src.DeepCopy()
			arr.RemoveByPath(tc.path)
			assert.Equal(t, tc.expected, arr)
		})
	}
}

func TestGetByPath(t *testing.T) {
	t.Parallel()

	doc := must.NotFail(NewDocument(
		"ismaster", true,
		"client", must.NotFail(NewDocument(
			"driver", must.NotFail(NewDocument(
				"name", "nodejs",
				"version", "4.0.0-beta.6",
			)),
			"os", must.NotFail(NewDocument(
				"type", "Darwin",
				"name", "darwin",
				"architecture", "x64",
				"version", "20.6.0",
			)),
			"platform", "Node.js v14.17.3, LE (unified)|Node.js v14.17.3, LE (unified)",
			"application", must.NotFail(NewDocument(
				"name", "mongosh 1.0.1",
			)),
		)),
		"compression", must.NotFail(NewArray("none")),
		"loadBalanced", false,
	))

	type testCase struct {
		path Path
		res  any
		err  string
	}

	for _, tc := range []testCase{{ //nolint:paralleltest // false positive
		path: NewStaticPath("compression", "0"),
		res:  "none",
	}, {
		path: NewStaticPath("compression"),
		res:  must.NotFail(NewArray("none")),
	}, {
		path: NewStaticPath("client", "driver"),
		res: must.NotFail(NewDocument(
			"name", "nodejs",
			"version", "4.0.0-beta.6",
		)),
	}, {
		path: NewStaticPath("client", "0"),
		err:  `types.getByPath: types.Document.Get: key not found: "0"`,
	}, {
		path: NewStaticPath("compression", "invalid"),
		err:  `types.getByPath: strconv.Atoi: parsing "invalid": invalid syntax`,
	}, {
		path: NewStaticPath("client", "missing"),
		err:  `types.getByPath: types.Document.Get: key not found: "missing"`,
	}, {
		path: NewStaticPath("compression", "1"),
		err:  `types.getByPath: types.Array.Get: index 1 is out of bounds [0-1)`,
	}, {
		path: NewStaticPath("compression", "0", "invalid"),
		err:  `types.getByPath: can't access string by path "invalid"`,
	}} {
		tc := tc
		t.Run(fmt.Sprint(tc.path), func(t *testing.T) {
			t.Parallel()

			res, err := getByPath(doc, tc.path)
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

func TestPathTrimSuffixPrefix(t *testing.T) {
	t.Parallel()

	pathOneElement := NewStaticPath("1")
	pathZeroElement := Path{s: []string{}}

	type testCase struct {
		name string
		f    func() Path
	}

	for _, tc := range []testCase{{
		name: "prefixOne",
		f:    pathOneElement.TrimPrefix,
	}, {
		name: "suffixOne",
		f:    pathOneElement.TrimSuffix,
	}, {
		name: "prefixZero",
		f:    pathZeroElement.TrimPrefix,
	}, {
		name: "suffixZero",
		f:    pathZeroElement.TrimSuffix,
	}} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Panics(t, func() {
				tc.f()
			})
		})
	}
}

func TestPathSuffixPrefix(t *testing.T) {
	t.Parallel()

	pathOneElement := NewStaticPath("1")
	pathZeroElement := Path{s: []string{}}

	type testCase struct {
		name string
		f    func() string
	}

	for _, tc := range []testCase{{
		name: "prefixOne",
		f:    pathOneElement.Prefix,
	}, {
		name: "suffixOne",
		f:    pathOneElement.Suffix,
	}, {
		name: "prefixZero",
		f:    pathZeroElement.Prefix,
	}, {
		name: "suffixZero",
		f:    pathZeroElement.Suffix,
	}} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Panics(t, func() {
				tc.f()
			})
		})
	}
}

func TestNewPathFromStringErrors(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		path string
		err  string
	}

	for _, tc := range []testCase{{
		name: "negative index",
		path: "foo.-1.bar",
		err:  "index must be greater than or equal to 0",
	}, {
		name: "empty path",
		path: "",
		err:  "path element must not be empty",
	}} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewPathFromString(tc.path)

			require.EqualError(t, err, tc.err)
		})
	}
}
