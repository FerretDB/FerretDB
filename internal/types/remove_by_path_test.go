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

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/stretchr/testify/assert"
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

	type testCase struct {
		name string
		path []string
		res  *Document
	}
	for _, tc := range []testCase{ //nolint:paralleltest // false positive
		{
			name: "test deep removal",
			path: []string{"client", "0", "foo", "baz", "baz", "baz"},
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
		{
			name: "not found no error, ismaster field removed",
			path: []string{"ismaster", "0"},
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
		{
			name: "removed entire client field",
			path: []string{"client"},
			res: must.NotFail(NewDocument(
				"ismaster", true,
				"value", must.NotFail(NewArray("none")),
			)),
		},
		{
			name: "only 1d array element of client field is removed",
			path: []string{"client", "1"},
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
		{
			name: "not found, element must be on place, no error",
			path: []string{"client", "3"},
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
		{
			name: "not found, no error doc is same",
			path: []string{"compression", "invalid"},
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
	} {
		tc := tc
		t.Run(fmt.Sprint(tc.path), func(t *testing.T) {
			t.Parallel()

			doc := sourceDoc.DeepCopy()
			doc.RemoveByPath(tc.path...)
			assert.Equal(t, tc.res, doc, tc.name)
		})
	}
}
