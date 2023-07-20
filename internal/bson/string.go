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

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// stringType represents BSON UTF-8 string type.
type stringType string

func (str *stringType) bsontype() {}

// readNested implements bsontype interface.
func (str *stringType) readNested(_ *bufio.Reader, _ int) error { return nil }

// ReadFrom implements bsontype interface.
func (str *stringType) ReadFrom(r *bufio.Reader) error {
	var l int32
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return lazyerrors.Error(err)
	}
	if l <= 0 {
		return lazyerrors.Errorf("invalid length %d", l)
	}

	b := make([]byte, l)
	if n, err := io.ReadFull(r, b); err != nil {
		return lazyerrors.Errorf("expected %d, read %d: %w", len(b), n, err)
	}

	if b[l-1] != 0 {
		return lazyerrors.Errorf("unexpected terminating byte %#02x", b[l-1])
	}

	*str = stringType(b[:l-1])
	return nil
}

// WriteTo implements bsontype interface.
func (str stringType) WriteTo(w *bufio.Writer) error {
	v, err := str.MarshalBinary()
	if err != nil {
		return lazyerrors.Error(err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (str stringType) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, int32(len(str)+1))
	buf.Write([]byte(str))
	buf.WriteByte(0)

	return buf.Bytes(), nil
}

// check interfaces
var (
	_ bsontype = (*stringType)(nil)
)
