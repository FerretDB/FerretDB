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

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// TestGetDocumentsAtSuffix demonstrates how getDocumentsAtSuffix works.
// The actual cases must be covered by integration tests to ensure compatibility.
func TestGetDocumentsAtSuffix(t *testing.T) {
	doc := must.NotFail(types.NewDocument(
		"v", must.NotFail(types.NewArray(
			must.NotFail(types.NewDocument("foo", must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument("bar", "hello")),
				must.NotFail(types.NewDocument("bar", "hohoho")),
			)))),
			must.NotFail(types.NewDocument("foo", must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument("bar", "world")),
			)))),
		))))

	for name, tc := range map[string]struct {
		filterKey  string
		wantSuffix string
		wantDocs   []*types.Document
	}{
		"dot-notation-indexes": {
			filterKey:  "v.0.foo.1.bar",
			wantSuffix: "bar",
			wantDocs: []*types.Document{
				must.NotFail(types.NewDocument("bar", "hohoho")),
			},
		},
		"dot-notation-no-index": {
			filterKey:  "v.foo.bar",
			wantSuffix: "bar",
			wantDocs: []*types.Document{
				must.NotFail(types.NewDocument("bar", "hello")),
				must.NotFail(types.NewDocument("bar", "hohoho")),
				must.NotFail(types.NewDocument("bar", "world")),
			},
		},
		"non-existent": {
			filterKey:  "a.b",
			wantSuffix: "b",
			wantDocs:   []*types.Document{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			path, err := types.NewPathFromString(tc.filterKey)
			assert.NoError(t, err)

			suffix, docs := getDocumentsAtSuffix(doc, path)

			assert.Equal(t, tc.wantSuffix, suffix)
			assert.Equal(t, tc.wantDocs, docs)
		})
	}
}
