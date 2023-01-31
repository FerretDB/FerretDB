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

package testutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestSetByPath(t *testing.T) {
	t.Parallel()

	newDoc := func() *types.Document {
		return must.NotFail(types.NewDocument(
			"client", must.NotFail(types.NewDocument(
				"driver", must.NotFail(types.NewDocument(
					"name", "nodejs",
				)),
			)),
			"compression", must.NotFail(types.NewArray("none")),
		))
	}

	type testCase struct {
		path  types.Path
		value any
		res   any
	}

	for _, tc := range []testCase{{ //nolint:paralleltest // false positive
		path:  types.NewPath("compression", "0"),
		value: "zstd",
		res: must.NotFail(types.NewDocument(
			"client", must.NotFail(types.NewDocument(
				"driver", must.NotFail(types.NewDocument(
					"name", "nodejs",
				)),
			)),
			"compression", must.NotFail(types.NewArray("zstd")),
		)),
	}, {
		path:  types.NewPath("client"),
		value: "foo",
		res: must.NotFail(types.NewDocument(
			"client", "foo",
			"compression", must.NotFail(types.NewArray("none")),
		)),
	}} {
		tc := tc
		t.Run(fmt.Sprint(tc.path), func(t *testing.T) {
			t.Parallel()

			doc := newDoc()
			SetByPath(t, doc, tc.value, tc.path)
			assert.Equal(t, tc.res, doc)
			assert.Equal(t, tc.value, GetExactByPath(t, doc, tc.path))
		})
	}
}
