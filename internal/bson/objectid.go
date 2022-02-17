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
	"io"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// objectIDType represents BSON ObjectId type.
type objectIDType types.ObjectID

func (obj *objectIDType) bsontype() {}

// ReadFrom implements bsontype interface.
func (obj *objectIDType) ReadFrom(r *bufio.Reader) error {
	if _, err := io.ReadFull(r, obj[:]); err != nil {
		return lazyerrors.Errorf("bson.ObjectID.ReadFrom (io.ReadFull): %w", err)
	}

	return nil
}

// WriteTo implements bsontype interface.
func (obj objectIDType) WriteTo(w *bufio.Writer) error {
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

// MarshalBinary implements bsontype interface.
func (obj objectIDType) MarshalBinary() ([]byte, error) {
	b := make([]byte, len(obj))
	copy(b, obj[:])
	return b, nil
}

// check interfaces
var (
	_ bsontype = (*objectIDType)(nil)
)
