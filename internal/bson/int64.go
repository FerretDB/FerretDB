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

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// int64Type represents BSON 64-bit integer type.
type int64Type int64

func (i *int64Type) bsontype() {}

// ReadFrom implements bsontype interface.
func (i *int64Type) ReadFrom(r *bufio.Reader, _ int) error {
	if err := binary.Read(r, binary.LittleEndian, i); err != nil {
		return lazyerrors.Errorf("bson.Int64.ReadFrom (binary.Read): %w", err)
	}

	return nil
}

// WriteTo implements bsontype interface.
func (i int64Type) WriteTo(w *bufio.Writer) error {
	v, err := i.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Int64.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Int64.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (i int64Type) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, i)

	return buf.Bytes(), nil
}

// check interfaces
var (
	_ bsontype = (*int64Type)(nil)
)
