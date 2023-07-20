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
	"math"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// doubleType represents BSON 64-bit binary floating point type.
type doubleType float64

func (d *doubleType) bsontype() {}

// readNested implements bsontype interface.
func (d *doubleType) readNested(_ *bufio.Reader, _ int) error { return nil }

// ReadFrom implements bsontype interface.
func (d *doubleType) ReadFrom(r *bufio.Reader) error {
	var bits uint64
	if err := binary.Read(r, binary.LittleEndian, &bits); err != nil {
		return lazyerrors.Errorf("bson.Double.ReadFrom (binary.Read): %w", err)
	}

	*d = doubleType(math.Float64frombits(bits))
	return nil
}

// WriteTo implements bsontype interface.
func (d doubleType) WriteTo(w *bufio.Writer) error {
	v, err := d.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Double.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Double.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (d doubleType) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, math.Float64bits(float64(d)))

	return buf.Bytes(), nil
}

// check interfaces
var (
	_ bsontype = (*doubleType)(nil)
)
