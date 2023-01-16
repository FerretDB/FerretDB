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
	"fmt"
	"strconv"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// arrayType represents BSON Array type.
type arrayType types.Array

func (a *arrayType) bsontype() {}

// ReadFrom implements bsontype interface.
func (a *arrayType) ReadFrom(r *bufio.Reader) error {
	var doc Document
	if err := doc.ReadFrom(r); err != nil {
		return lazyerrors.Error(err)
	}

	keys := doc.Keys()
	values := doc.Values()

	if len(keys) != len(values) {
		panic(fmt.Sprintf("document must have the same number of keys and values (keys: %d, values: %d)", len(keys), len(values)))
	}

	ta := types.MakeArray(len(keys))

	for i, k := range keys {
		if k != strconv.Itoa(i) {
			return lazyerrors.Errorf("key %d is %q", i, k)
		}

		v := values[i]

		ta.Append(v)
	}

	*a = arrayType(*ta)
	return nil
}

// WriteTo implements bsontype interface.
func (a arrayType) WriteTo(w *bufio.Writer) error {
	v, err := a.MarshalBinary()
	if err != nil {
		return lazyerrors.Error(err)
	}

	if _, err = w.Write(v); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (a arrayType) MarshalBinary() ([]byte, error) {
	ta := types.Array(a)
	l := ta.Len()

	fields := make([]field, l)
	for i := 0; i < l; i++ {
		key := strconv.Itoa(i)
		value, err := ta.Get(i)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		fields[i] = field{key: key, value: value}
	}

	doc := Document{
		fields: fields,
	}

	b, err := doc.MarshalBinary()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return b, nil
}

// check interfaces
var (
	_ bsontype = (*arrayType)(nil)
)
