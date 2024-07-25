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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
)

// Array represents a BSON array in the (partially) decoded form.
type Array struct {
	elements        []any
	frozen          bool
	*wirebson.Array // embed to delegate method
}

// NewArray creates a new Array from the given values.
func NewArray(values ...any) (*Array, error) {
	arr, err := wirebson.NewArray(values)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Array{Array: arr}, nil
}

// MakeArray creates a new empty Array with the given capacity.
func MakeArray(cap int) *Array {
	return &Array{
		Array: wirebson.MakeArray(cap),
	}
}

// TypesArray gets an array, decodes and converts to [*types.Array].
func TypesArray(arr wirebson.AnyArray) (*types.Array, error) {
	wArr, err := arr.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	bArr := &Array{Array: wArr}

	tArr, err := bArr.Convert()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return tArr, nil
}

// Freeze prevents array from further modifications.
// Any methods that would modify the array will panic.
//
// It is safe to call Freeze multiple times.
func (arr *Array) Freeze() {
	arr.Array.Freeze()
}

// Len returns the number of elements in the Array.
func (arr *Array) Len() int {
	return arr.Array.Len()
}

// Get returns the element at the given index.
// It panics if index is out of bounds.
func (arr *Array) Get(index int) any {
	return arr.Array.Get(index)
}

// Add adds a new element to the Array.
func (arr *Array) Add(value any) error {
	switch v := value.(type) {
	case *Document:
		value = v.Document
	case *Array:
		value = v.Array
	}

	return arr.Array.Add(value)
}

// Replace sets the value of the element at the given index.
// It panics if index is out of bounds.
func (arr *Array) Replace(index int, value any) error {
	return arr.Array.Replace(index, value)
}
