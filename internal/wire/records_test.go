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
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// loadRecords gets recursively all .bin files from the recordsPath directory,
// parses their content to wire Msgs and returns them as an array of testCase
// structs with headerB and bodyB fields set.
// If no records are found, it returns nil and no error.
func loadRecords(recordsPath string) ([]testCase, error) {
	// Load recursively every file path with ".bin" extension from recordsPath directory
	var recordFiles []string

	err := filepath.WalkDir(recordsPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(entry.Name()) == ".bin" {
			recordFiles = append(recordFiles, path)
		}

		return nil
	})

	switch {
	case os.IsNotExist(err):
		return nil, nil
	case err != nil:
		return nil, err
	}

	// Select random 1000 number of files from an array of files
	if len(recordFiles) > 1000 {
		idx := rand.Perm(len(recordFiles))
		idx = idx[:1000]

		sel := []string{}
		for _, i := range idx {
			sel = append(sel, recordFiles[i])
		}
		recordFiles = sel
	}

	var resMsgs []testCase

	// Read every record file, parse their content to wire messages
	// and store them in the testCase struct
	for _, path := range recordFiles {
		f, err := os.Open(path)
		if err != nil {
			return nil, lazyerrors.Errorf("%s: %w", path, err)
		}

		defer f.Close()

		r := bufio.NewReader(f)

		for {
			header, body, err := ReadMessage(r)
			if errors.Is(err, ErrZeroRead) {
				break
			}
			if err != nil {
				return nil, lazyerrors.Errorf("%s: %w", path, err)
			}

			headBytes, err := header.MarshalBinary()
			if err != nil {
				return nil, lazyerrors.Errorf("%s: %w", path, err)
			}

			bodyBytes, err := body.MarshalBinary()
			if err != nil {
				return nil, lazyerrors.Errorf("%s: %w", path, err)
			}

			resMsgs = append(resMsgs, testCase{
				headerB: headBytes,
				bodyB:   bodyBytes,
			})
		}
	}

	return resMsgs, nil
}
