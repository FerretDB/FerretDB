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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
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

// LoadDataDocuments is like LoadRecords, but instead of loading raw records,
// it extracts valid data documents.
func LoadDataDocuments(dir string, limit int) ([]*types.Document, error) {
	files, err := findBinFiles(dir)
	if err != nil {
		return nil, err
	}

	var docs []*types.Document

	// We can't tell whether a record have any data documents or
	// not before parsing the bin file.
	// Parsing all files and then using some shuffling is slow,
	// so we're using a random window to collect the results.

	i := 0

	if limit == 0 {
		limit = len(files)
	} else {
		i = rand.Intn(len(files))
	}
	for j := 0; j < len(files); j, i = j+1, i+1 {
		if i > len(files) {
			i = 0 // wrap around
		}
		f := files[i]

		recs, err := loadRecordFile(f)
		if err != nil {
			return nil, lazyerrors.Errorf("%s: %w", f, err)
		}

		for i, r := range recs {
			rDocs, err := extractRecordDataDocuments(r)
			if err != nil {
				return nil, lazyerrors.Errorf("%s record[%d]: %w", f, i, err)
			}

			docs = append(docs, rDocs...)
			if len(docs) >= limit {
				docs = docs[:limit]
				return docs, nil
			}
		}
	}

	return docs, nil
}

// LoadRecords finds all .bin files recursively, selects up to the limit at random (or all if limit <= 0), and parses them.
func LoadRecords(dir string, limit int) ([]Record, error) {
	files, err := findBinFiles(dir)
	if err != nil {
		return nil, err
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

// findBinFiles collects the filenames inside dir that have ".bin" suffix.
func findBinFiles(dir string) ([]string, error) {
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
	default:
		return files, nil
	}
}

// extractRecordDataDocuments collects all data documents from a given record.
func extractRecordDataDocuments(r Record) ([]*types.Document, error) {
	var docs []*types.Document

	switch b := r.Body.(type) {
	case *OpMsg:
		doc, err := b.Document()
		if err != nil {
			return nil, err
		}
		docs, err = appendDataDocuments(docs, doc)
		if err != nil {
			return nil, lazyerrors.Errorf("OpMsg: %w", err)
		}

	case *OpReply:
		var err error
		for i, d := range b.Documents {
			docs, err = appendDataDocuments(docs, d)
			if err != nil {
				return nil, lazyerrors.Errorf("OpReply.documents[%d]: %w", i, err)
			}
		}
		docs = append(docs, b.Documents...)
	}

	return docs, nil
}

// appendDataDocuments finds all valid data documents in v and pushes them to dst.
func appendDataDocuments(dst []*types.Document, v any) ([]*types.Document, error) {
	var err error

	switch v := v.(type) {
	case *types.Document:
		if v.ValidateData() == nil {
			dst = append(dst, v)
			return dst, nil
		}

		if v.Has("insert") && v.Has("documents") {
			dst, err = appendDataDocuments(dst, must.NotFail(v.Get("documents")))
			if err != nil {
				return nil, lazyerrors.Errorf("insert.documents: %w", err)
			}
		}

	case *types.Array:
		for i := 0; i < v.Len(); i++ {
			dst, err = appendDataDocuments(dst, must.NotFail(v.Get(i)))
			if err != nil {
				return nil, lazyerrors.Errorf("[%d]: %w", i, err)
			}
		}
	}

	return dst, nil
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

		// TODO we should still set HeaderB and BodyB
		// https://github.com/FerretDB/FerretDB/issues/1636
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
