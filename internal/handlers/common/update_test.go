package common

import (
	"fmt"
	"testing"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/stretchr/testify/require"
)

func TestValidatorTree(t *testing.T) {
	t.Parallel()

	for i, tc := range []struct {
		paths   []string
		noError bool
		errOn   int
	}{
		{paths: []string{"v.foo.bar", "v.bar.foo", "v.foo"}, errOn: 2},
		{paths: []string{"v.foo", "v.foo.bar"}, errOn: 1},
		{paths: []string{"v.foo", "v.bar"}, noError: true},
	} {
		i, tc := i, tc
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tree := newValidatorTree()

			for i, p := range tc.paths {
				path, err := types.NewPathFromString(p)
				require.NoError(t, err)

				err = tree.Validate(path)
				if tc.noError {
					require.NoError(t, err)
					continue
				}

				if tc.errOn == i {
					require.Error(t, err, "Error expected at index %d", tc.errOn)
				} else {
					require.NoError(t, err, "Error expected at index %d, but occurred at %d", tc.errOn, i)
				}
			}
		})

	}
}
