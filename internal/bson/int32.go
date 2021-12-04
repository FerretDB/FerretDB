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
	"encoding/json"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Int32 represents BSON Int32 data type.
type Int32 int32

func (i *Int32) bsontype() {}

// ReadFrom implements bsontype interface.
func (i *Int32) ReadFrom(r *bufio.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, i); err != nil {
		return lazyerrors.Errorf("bson.Int32.ReadFrom (binary.Read): %w", err)
	}

	return nil
}

// WriteTo implements bsontype interface.
func (i Int32) WriteTo(w *bufio.Writer) error {
	v, err := i.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Int32.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Int32.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (i Int32) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, i)

	return buf.Bytes(), nil
}

// UnmarshalJSON implements bsontype interface.
func (i *Int32) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o int32
	if err := dec.Decode(&o); err != nil {
		return err
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Errorf("bson.Int32.UnmarshalJSON: %s", err)
	}

	*i = Int32(o)
	return nil
}

// MarshalJSON implements bsontype interface.
func (i Int32) MarshalJSON() ([]byte, error) {
	return json.Marshal(int32(i))
}

// check interfaces
var (
	_ bsontype = (*Int32)(nil)
)
