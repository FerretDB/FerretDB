package common

import (
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_walkProjectionPath(t *testing.T) {
	testCases := map[string]struct {
		path      types.Path
		inclusion bool
		projected *types.Document
		doc       *types.Document
		expected  *types.Document
	}{
		"Inclusion2Level2": {
			path:      types.NewStaticPath("a", "b"),
			inclusion: true,
			projected: must.NotFail(types.NewDocument("_id", "level2")),
			doc:       must.NotFail(types.NewDocument("_id", "level2", "a", must.NotFail(types.NewDocument("b", 12)))),
			expected:  must.NotFail(types.NewDocument("_id", "level2", "a", must.NotFail(types.NewDocument("b", 12)))),
		},
		"Inclusion3Level3": {
			path:      types.NewStaticPath("a", "b", "c"),
			inclusion: true,
			projected: must.NotFail(types.NewDocument("_id", "level3")),
			doc:       must.NotFail(types.NewDocument("_id", "level3", "a", must.NotFail(types.NewDocument("b", must.NotFail(types.NewDocument("c", 12)))))),
			expected:  must.NotFail(types.NewDocument("_id", "level3", "a", must.NotFail(types.NewDocument("b", must.NotFail(types.NewDocument("c", 12)))))),
		},
		"Inclusion3Level2": {
			path:      types.NewStaticPath("a", "b", "c"),
			inclusion: true,
			projected: must.NotFail(types.NewDocument("_id", "level2")),
			doc:       must.NotFail(types.NewDocument("_id", "level2", "a", must.NotFail(types.NewDocument("b", 12)))),
			expected:  must.NotFail(types.NewDocument("_id", "level2", "a", types.MakeDocument(0))),
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			res := walkProjectionPath(tc.path, tc.inclusion, tc.projected, tc.doc)

			assert.Equal(t, tc.expected, res)
		})
	}
}
