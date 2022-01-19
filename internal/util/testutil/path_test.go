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
		return types.MustNewDocument(
			"client", types.MustNewDocument(
				"driver", types.MustNewDocument(
					"name", "nodejs",
				),
			),
			"compression", must.NotFail(types.NewArray("none")),
		)
	}

	type testCase struct {
		path  []string
		value any
		res   any
	}

	for _, tc := range []testCase{{ //nolint:paralleltest // false positive
		path:  []string{"compression", "0"},
		value: "zstd",
		res: types.MustNewDocument(
			"client", types.MustNewDocument(
				"driver", types.MustNewDocument(
					"name", "nodejs",
				),
			),
			"compression", must.NotFail(types.NewArray("zstd")),
		),
	}, {
		path:  []string{"client"},
		value: "foo",
		res: types.MustNewDocument(
			"client", "foo",
			"compression", must.NotFail(types.NewArray("none")),
		),
	}} {
		tc := tc
		t.Run(fmt.Sprint(tc.path), func(t *testing.T) {
			t.Parallel()

			doc := newDoc()
			SetByPath(t, doc, tc.value, tc.path...)
			assert.Equal(t, tc.res, doc)
			assert.Equal(t, tc.value, GetByPath(t, doc, tc.path...))
		})
	}
}
