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

package postgresql

import (
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// scanner scans one or more columns from rows and returns *types.Document.
type scanner interface {
	Scan(rows pgx.Rows) (*types.Document, error)
}

// documentScanner scans a single bytes column of rows.
type documentScanner struct{}

// Scan scans bytes from rows and unmarshals bytes to *types.Document.
func (s *documentScanner) Scan(rows pgx.Rows) (*types.Document, error) {
	var b []byte
	if err := rows.Scan(&b); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := sjson.Unmarshal(b)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// recordIDScanner scans recordID column and bytes column.
type recordIDScanner struct{}

// Scan scans recordID and bytes from rows and unmarshals bytes to *types.Document
// then sets recordID to the document.
func (s *recordIDScanner) Scan(rows pgx.Rows) (*types.Document, error) {
	var recordID types.Timestamp
	var b []byte

	if err := rows.Scan(&recordID, &b); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := sjson.Unmarshal(b)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc.SetRecordID(recordID)

	return doc, nil
}

// onlyRecordIDScanner scans recordID column only.
type onlyRecordIDScanner struct{}

// Scan scans recordID from rows and creates new *types.Document with recordID.
func (s *onlyRecordIDScanner) Scan(rows pgx.Rows) (*types.Document, error) {
	var recordID types.Timestamp

	if err := rows.Scan(&recordID); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc := must.NotFail(types.NewDocument())
	doc.SetRecordID(recordID)

	return doc, nil
}
