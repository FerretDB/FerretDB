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

package wire

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadRecords reads all .bin files from the given directory,
// parses their content to wire messages and returns them as test cases.
func loadRecords(tb testing.TB, dir string) []testCase {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	var res []testCase

	// walk directory recursively in case we want to change its layout
	err := filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		require.NoError(tb, err)

		if filepath.Ext(path) != ".bin" {
			return nil
		}

		f, err := os.Open(path)
		require.NoError(tb, err)
		defer f.Close()

		bufr := bufio.NewReader(f)

		for {
			msgHeader, msgBody, err := ReadMessage(bufr)

			if errors.Is(err, io.EOF) {
				assert.Zero(tb, bufr.Buffered(), "not all bufr bytes were consumed")
				return nil
			}
			require.NoError(tb, err)

			headerB, err := msgHeader.MarshalBinary()
			require.NoError(tb, err)

			bodyB, err := msgBody.MarshalBinary()
			require.NoError(tb, err)

			res = append(res, testCase{
				name:    filepath.Base(path),
				headerB: headerB,
				bodyB:   bodyB,
				header:  msgHeader,
				body:    msgBody,
			})
		}
	})

	require.NoError(tb, err)

	return res
}

func TestRecords(t *testing.T) {
	t.Skip("TODO") // HACK

	t.Parallel()
	testMessages(t, loadRecords(t, filepath.Join("..", "..", "records")))
}
