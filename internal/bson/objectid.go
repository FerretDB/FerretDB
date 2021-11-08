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
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

type ObjectID [12]byte

func (obj *ObjectID) bsontype() {}

func (obj *ObjectID) ReadFrom(r *bufio.Reader) error {
	if _, err := io.ReadFull(r, obj[:]); err != nil {
		return lazyerrors.Errorf("bson.ObjectID.ReadFrom (io.ReadFull): %w", err)
	}

	return nil
}

func (obj ObjectID) WriteTo(w *bufio.Writer) error {
	v, err := obj.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.ObjectID.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.ObjectID.WriteTo: %w", err)
	}

	return nil
}

func (obj ObjectID) MarshalBinary() ([]byte, error) {
	b := make([]byte, len(obj))
	copy(b, obj[:])
	return b, nil
}

type objectIDJSON struct {
	O string `json:"$o"`
}

func (obj *ObjectID) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o objectIDJSON
	if err := dec.Decode(&o); err != nil {
		return err
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Errorf("bson.ObjectID.UnmarshalJSON: %s", err)
	}

	b, err := hex.DecodeString(o.O)
	if err != nil {
		return err
	}
	if len(b) != 12 {
		return lazyerrors.Errorf("bson.ObjectID.UnmarshalJSON: %d bytes", len(b))
	}
	copy(obj[:], b)

	return nil
}

func (obj ObjectID) MarshalJSON() ([]byte, error) {
	return json.Marshal(objectIDJSON{
		O: hex.EncodeToString(obj[:]),
	})
}

// check interfaces
var (
	_ bsontype = (*ObjectID)(nil)
)
