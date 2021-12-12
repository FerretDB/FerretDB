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
	"time"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// DateTime represents BSON DateTime data type.
type DateTime time.Time

func (dt DateTime) String() string {
	return time.Time(dt).Format(time.RFC3339Nano)
}

func (dt *DateTime) bsontype() {}

// ReadFrom implements bsontype interface.
func (dt *DateTime) ReadFrom(r *bufio.Reader) error {
	var ts int64
	if err := binary.Read(r, binary.LittleEndian, &ts); err != nil {
		return lazyerrors.Errorf("bson.DateTime.ReadFrom (binary.Read): %w", err)
	}

	// TODO Use .UTC(): https://github.com/FerretDB/FerretDB/issues/43
	*dt = DateTime(time.UnixMilli(ts))
	return nil
}

// WriteTo implements bsontype interface.
func (dt DateTime) WriteTo(w *bufio.Writer) error {
	v, err := dt.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.DateTime.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.DateTime.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (dt DateTime) MarshalBinary() ([]byte, error) {
	ts := time.Time(dt).UnixMilli()

	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, ts)

	return buf.Bytes(), nil
}

type dateTimeJSON struct {
	D int64 `json:"$d"`
}

// UnmarshalJSON implements bsontype interface.
func (dt *DateTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o dateTimeJSON
	if err := dec.Decode(&o); err != nil {
		return err
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Errorf("bson.DateTime.UnmarshalJSON: %s", err)
	}

	// TODO Use .UTC(): https://github.com/FerretDB/FerretDB/issues/43
	*dt = DateTime(time.UnixMilli(o.D))
	return nil
}

// MarshalJSON implements bsontype interface.
func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(dateTimeJSON{
		D: time.Time(dt).UnixMilli(),
	})
}

// check interfaces
var (
	_ bsontype = (*DateTime)(nil)
)
