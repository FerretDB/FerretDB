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

// Record represents a single recorded wire protocol message, loaded from a .bin file.
type Record struct {
	// those may be unset if message is invalid
	Header *MsgHeader
	Body   MsgBody

	// those are always set
	HeaderB []byte
	BodyB   []byte
}

// LoadRecords finds all .bin files recursively, selects up to the limit at random (or all if limit <= 0), and parses them.
func LoadRecords(dir string, limit int) ([]Record, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return lazyerrors.Error(err)
		}

		if filepath.Ext(entry.Name()) == ".bin" {
			files = append(files, path)
		}

		return nil
	})

	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case err != nil:
		return nil, lazyerrors.Error(err)
	}

	if limit > 0 && len(files) > limit {
		f := make([]string, limit)
		for fI, filesI := range rand.Perm(len(files))[:limit] {
			f[fI] = files[filesI]
		}
		files = f
	}

	var res []Record

	for _, file := range files {
		r, err := loadRecordFile(file)
		if err != nil {
			return nil, lazyerrors.Errorf("%s: %w", file, err)
		}

		res = append(res, r...)
	}

	return res, nil
}

// loadRecordFile parses a single .bin file.
func loadRecordFile(file string) ([]Record, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer f.Close() //nolint:errcheck // we are only reading it

	r := bufio.NewReader(f)

	var res []Record

	for {
		header, body, err := ReadMessage(r)
		if errors.Is(err, ErrZeroRead) {
			break
		}

		// we should still set HeaderB and BodyB
		// TODO https://github.com/FerretDB/FerretDB/issues/1636
		if err != nil {
			break
		}

		headerB, err := header.MarshalBinary()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		bodyB, err := body.MarshalBinary()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res = append(res, Record{
			Header:  header,
			Body:    body,
			HeaderB: headerB,
			BodyB:   bodyB,
		})
	}

	return res, nil
}
