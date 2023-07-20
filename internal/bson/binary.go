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
	"io"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// binaryType represents BSON Binary data type.
type binaryType types.Binary

func (bin *binaryType) bsontype() {}

// ReadFrom implements bsontype interface.
func (bin *binaryType) ReadFrom(r *bufio.Reader) error {
	var l int32
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return lazyerrors.Errorf("bson.Binary.ReadFrom (binary.Read): %w", err)
	}
	if l < 0 {
		return lazyerrors.Errorf("bson.Binary.ReadFrom: invalid length: %d", l)
	}

	subtype, err := r.ReadByte()
	if err != nil {
		return lazyerrors.Errorf("bson.Binary.ReadFrom (ReadByte): %w", err)
	}
	bin.Subtype = types.BinarySubtype(subtype)

	bin.B = make([]byte, l)
	if _, err := io.ReadFull(r, bin.B); err != nil {
		return lazyerrors.Errorf("bson.Binary.ReadFrom (io.ReadFull): %w", err)
	}

	return nil
}

// WriteTo implements bsontype interface.
func (bin binaryType) WriteTo(w *bufio.Writer) error {
	v, err := bin.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Binary.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Binary.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (bin binaryType) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, int32(len(bin.B)))
	buf.WriteByte(byte(bin.Subtype))
	buf.Write(bin.B)

	return buf.Bytes(), nil
}

// check interfaces
var (
	_ bsontype = (*binaryType)(nil)
)
