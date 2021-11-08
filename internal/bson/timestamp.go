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

	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

type Timestamp uint64

func (ts *Timestamp) bsontype() {}

func (ts *Timestamp) ReadFrom(r *bufio.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, ts); err != nil {
		return lazyerrors.Errorf("bson.Timestamp.ReadFrom (binary.Read): %w", err)
	}

	return nil
}

func (ts Timestamp) WriteTo(w *bufio.Writer) error {
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

func (ts Timestamp) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, ts)

	return buf.Bytes(), nil
}

type timestampJSON struct {
	T uint64 `json:"$t,string"`
}

func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o timestampJSON
	if err := dec.Decode(&o); err != nil {
		return err
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Errorf("bson.Timestamp.UnmarshalJSON: %s", err)
	}

	*ts = Timestamp(o.T)
	return nil
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(timestampJSON{
		T: uint64(ts),
	})
}

// check interfaces
var (
	_ bsontype = (*Timestamp)(nil)
)
