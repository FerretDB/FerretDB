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

package clientconn

import (
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/FerretDB/wire"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestRenameFile(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	h := sha256.New()
	dir := t.TempDir()
	c := &conn{
		testRecordsDir: dir,
		l:              testutil.Logger(t),
	}

	t.Run("PartialFileDeleted", func(t *testing.T) {
		f, err := os.CreateTemp(dir, "_*.partial")
		require.NoError(t, err)

		c.renamePartialFile(ctx, f, h, errors.New("some error"))

		_, err = os.Stat(f.Name())
		require.True(t, os.IsNotExist(err))

		files, err := filepath.Glob(filepath.Join(dir, "*", "*.bin"))
		require.NoError(t, err)
		require.Empty(t, files)
	})

	t.Run("FileRenamed", func(t *testing.T) {
		f, err := os.CreateTemp(dir, "_*.partial")
		require.NoError(t, err)

		c.renamePartialFile(ctx, f, h, wire.ErrZeroRead)

		_, err = os.Stat(f.Name())
		require.True(t, os.IsNotExist(err))

		files, err := filepath.Glob(filepath.Join(dir, "*", "*.bin"))
		require.NoError(t, err)
		require.Len(t, files, 1)
	})
}
