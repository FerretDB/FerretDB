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

package bson

import (
	"bufio"
	"bytes"
	"encoding/binary"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// timestampType represents BSON Timestamp type.
type timestampType types.Timestamp

func (ts *timestampType) bsontype() {}

// readNested implements bsontype interface.
func (ts *timestampType) readNested(r *bufio.Reader, _ int) error { return nil }

// ReadFrom implements bsontype interface.
func (ts *timestampType) ReadFrom(r *bufio.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, ts); err != nil {
		return lazyerrors.Errorf("bson.Timestamp.ReadFrom (binary.Read): %w", err)
	}

	return nil
}

// WriteTo implements bsontype interface.
func (ts timestampType) WriteTo(w *bufio.Writer) error {
	v, err := ts.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Timestamp.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Timestamp.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (ts timestampType) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, ts)

	return buf.Bytes(), nil
}

// check interfaces
var (
	_ bsontype = (*timestampType)(nil)
)
