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

package tjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestUnmarshalTJSON(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:paralleltest // false positive
		in  driver.Document
		out *types.Document
		err error
	}{
		"EmptyJSON": {
			in:  driver.Document(`{}`),
			out: new(types.Document),
			err: nil,
		},
		"EmptyArray": {
			in: driver.Document(`{"emptyArray":[]}`),
			out: must.NotFail(types.NewDocument(
				"emptyArray", must.NotFail(types.NewArray([]any{}...)),
			)),
			err: nil,
		},
		"BoolStringNullDouble": {
			in: driver.Document(
				`{"_id": "string_type_id", "enabled": true, "float64": 42.13, "null_value": null}`),
			out: must.NotFail(types.NewDocument(
				"_id", "string_type_id",
				"enabled", true,
				"float64", float64(42.13),
				"null_value", types.Null,
			)),
			err: nil,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v, err := Unmarshal(tc.in)
			if !assert.Equal(t, tc.err, err, name) {
				t.Log(err)
			}

			if tc.err == nil {
				assert.Equal(t, tc.out, v, name)
			}
		})
	}
}
