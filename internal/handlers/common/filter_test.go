package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// TestFindLeavesForFilter demonstrates how findLeavesForFilter works.
// The actual cases must be covered by integration tests to ensure compatibility.
func TestFindLeavesForFilter(t *testing.T) {
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

	for _, tc := range map[string]struct {
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
		path, err := types.NewPathFromString("a.b")
		assert.NoError(t, err)

		suffix, docs := findLeavesForFilter(doc, path)

		assert.Equal(t, tc.wantSuffix, suffix)
		assert.Equal(t, tc.wantDocs, docs)
	}
}
