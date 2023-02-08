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

package tigrisdb

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tigrisdata/tigris-client-go/driver"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestBuildFilter(t *testing.T) {
	t.Parallel()
	tdb := TigrisDB{}

	for name, tc := range map[string]struct {
		filter   *types.Document
		expected string
		skip     string
	}{
		"String": {
			filter:   must.NotFail(types.NewDocument("v", "foo")),
			expected: `{"v":"foo"}`,
		},
		"EmptyString": {
			filter:   must.NotFail(types.NewDocument("v", "")),
			expected: `{"v":""}`,
			skip:     "https://github.com/FerretDB/FerretDB/issues/1940",
		},
		"Int32": {
			filter:   must.NotFail(types.NewDocument("v", int32(42))),
			expected: `{"v":42}`,
		},
		"Int64": {
			filter:   must.NotFail(types.NewDocument("v", int64(42))),
			expected: `{"v":42}`,
		},
		"Float64": {
			filter:   must.NotFail(types.NewDocument("v", float64(42.13))),
			expected: `{"v":42.13}`,
		},
		"MaxFloat64": {
			filter:   must.NotFail(types.NewDocument("v", math.MaxFloat64)),
			expected: `{"v":1.7976931348623157e+308}`,
		},
		"Bool": {
			filter: must.NotFail(types.NewDocument("v", true)),
		},
		"Comment": {
			filter: must.NotFail(types.NewDocument("$comment", "I'm comment")),
		},
		"ObjectID": {
			filter: must.NotFail(types.NewDocument("v",
				types.ObjectID(must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011"))),
			)),
			expected: `{"v":"AAECAwQFBgcICRAR"}`,
		},
		"IDObjectID": {
			filter: must.NotFail(types.NewDocument("_id", types.ObjectID(
				must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011")),
			))),
			expected: `{"_id":"AAECAwQFBgcICRAR"}`,
		},
		"IDString": {
			filter:   must.NotFail(types.NewDocument("_id", "foo")),
			expected: `{"_id":"foo"}`,
		},
		"IDDotNotation": {
			filter:   must.NotFail(types.NewDocument("_id.doc", "foo")),
			expected: `{"_id.doc":"foo"}`,
		},
		"DotNotation": {
			filter:   must.NotFail(types.NewDocument("v.doc", "foo")),
			expected: `{"v.doc":"foo"}`,
		},
		"DotNotationArrayIndex": {
			filter: must.NotFail(types.NewDocument("v.arr.0", "foo")),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			// replace default value with default json
			if tc.expected == "" {
				tc.expected = "{}"
			}

			expected := driver.Filter(tc.expected)
			actual := tdb.BuildFilter(tc.filter)

			assert.Equal(t, expected, actual)
		})
	}
}
