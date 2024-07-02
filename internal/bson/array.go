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
	"bytes"
	"encoding/binary"
	"log/slog"
	"strconv"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Array represents a BSON array in the (partially) decoded form.
type Array struct {
	elements []any
	frozen   bool
}

// NewArray creates a new Array from the given values.
func NewArray(values ...any) (*Array, error) {
	res := &Array{
		elements: make([]any, 0, len(values)),
	}

	for i, v := range values {
		if err := res.Add(v); err != nil {
			return nil, lazyerrors.Errorf("%d: %w", i, err)
		}
	}

	return res, nil
}

// MakeArray creates a new empty Array with the given capacity.
func MakeArray(cap int) *Array {
	return &Array{
		elements: make([]any, 0, cap),
	}
}

// Freeze prevents array from further modifications.
// Any methods that would modify the array will panic.
//
// It is safe to call Freeze multiple times.
func (arr *Array) Freeze() {
	arr.frozen = true
}

// checkFrozen panics if array is frozen.
func (arr *Array) checkFrozen() {
	if arr.frozen {
		panic("array is frozen and can't be modified")
	}
}

// Len returns the number of elements in the Array.
func (arr *Array) Len() int {
	return len(arr.elements)
}

// Get returns the element at the given index.
// It panics if index is out of bounds.
func (arr *Array) Get(index int) any {
	return arr.elements[index]
}

// Add adds a new element to the Array.
func (arr *Array) Add(value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Error(err)
	}

	arr.checkFrozen()

	arr.elements = append(arr.elements, value)

	return nil
}

// Replace sets the value of the element at the given index.
// It panics if index is out of bounds.
func (arr *Array) Replace(index int, value any) error {
	if err := validBSONType(value); err != nil {
		return lazyerrors.Error(err)
	}

	arr.checkFrozen()

	arr.elements[index] = value

	return nil
}

// Encode encodes BSON array.
//
// TODO https://github.com/FerretDB/FerretDB/issues/3759
// This method should accept a slice of bytes, not return it.
// That would allow to avoid unnecessary allocations.
func (arr *Array) Encode() (RawArray, error) {
	size := sizeAny(arr)
	buf := bytes.NewBuffer(make([]byte, 0, size))

	if err := binary.Write(buf, binary.LittleEndian, uint32(size)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	for i, v := range arr.elements {
		if err := encodeField(buf, strconv.Itoa(i), v); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	if err := binary.Write(buf, binary.LittleEndian, byte(0)); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return buf.Bytes(), nil
}

// Decode returns itself to implement [AnyArray].
//
// Receiver must not be nil.
func (arr *Array) Decode() (*Array, error) {
	must.NotBeZero(arr)
	return arr, nil
}

// LogValue implements [slog.LogValuer].
func (arr *Array) LogValue() slog.Value {
	return slogValue(arr, 1)
}

// check interfaces
var (
	_ AnyArray       = (*Array)(nil)
	_ slog.LogValuer = (*Array)(nil)
)
