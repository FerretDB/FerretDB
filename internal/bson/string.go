// Copyright 2021 Baltoro OÜ.
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
	"encoding/json"
	"io"

	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

type String string

func (str *String) bsontype() {}

func (str *String) ReadFrom(r *bufio.Reader) error {
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

	*str = String(b[:l-1])
	return nil
}

func (str String) WriteTo(w *bufio.Writer) error {
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

func (str String) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, int32(len(str)+1))
	buf.Write([]byte(str))
	buf.WriteByte(0)

	return buf.Bytes(), nil
}

func (str *String) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return lazyerrors.Error(err)
	}

	*str = String(s)
	return nil
}

func (str String) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(string(str))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}

// check interfaces
var (
	_ bsontype = (*String)(nil)
)
