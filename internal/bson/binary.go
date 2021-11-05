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

	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

type Binary struct {
	Subtype types.BinarySubtype
	B       []byte
}

func (bin *Binary) bsontype() {}

func (bin *Binary) ReadFrom(r *bufio.Reader) error {
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

func (bin Binary) WriteTo(w *bufio.Writer) error {
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

type binaryJSON struct {
	B []byte `json:"$b"`
	S byte   `json:"s"`
}

func (bin Binary) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, int32(len(bin.B)))
	buf.WriteByte(byte(bin.Subtype))
	buf.Write(bin.B)

	return buf.Bytes(), nil
}

func (bin *Binary) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o binaryJSON
	err := dec.Decode(&o)
	if err != nil {
		return lazyerrors.Error(err)
	}
	if err = checkConsumed(dec, r); err != nil {
		return lazyerrors.Errorf("bson.Binary.UnmarshalJSON: %w", err)
	}

	bin.B = o.B
	bin.Subtype = types.BinarySubtype(o.S)
	return nil
}

func (bin Binary) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(binaryJSON{
		B: bin.B,
		S: byte(bin.Subtype),
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return b, nil
}

// check interfaces
var (
	_ bsontype = (*Binary)(nil)
)
