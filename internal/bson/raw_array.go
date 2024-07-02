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
	"log/slog"
	"strconv"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// RawArray represents a single BSON array in the binary encoded form.
//
// It generally references a part of a larger slice, not a copy.
type RawArray []byte

// Encode returns itself to implement the [AnyArray] interface.
func (raw RawArray) Encode() (RawArray, error) {
	return raw, nil
}

// Decode decodes a single BSON array that takes the whole byte slice.
//
// Only top-level elements are decoded;
// nested documents and arrays are converted to RawDocument and RawArray respectively,
// using raw's subslices without copying.
func (raw RawArray) Decode() (*Array, error) {
	res, err := raw.decode(decodeShallow)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// DecodeDeep decodes a single valid BSON array that takes the whole byte slice.
//
// All nested documents and arrays are decoded recursively.
func (raw RawArray) DecodeDeep() (*Array, error) {
	res, err := raw.decode(decodeDeep)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// decode decodes a single BSON array that takes the whole byte slice.
func (raw RawArray) decode(mode decodeMode) (*Array, error) {
	doc, err := RawDocument(raw).decode(mode)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := &Array{
		elements: make([]any, len(doc.fields)),
	}

	for i, f := range doc.fields {
		if f.name != strconv.Itoa(i) {
			return nil, lazyerrors.Errorf("invalid array index: %q", f.name)
		}

		res.elements[i] = f.value
	}

	return res, nil
}

// LogValue implements [slog.LogValuer].
func (raw RawArray) LogValue() slog.Value {
	return slogValue(raw, 1)
}

// check interfaces
var (
	_ AnyArray       = RawArray(nil)
	_ slog.LogValuer = RawArray(nil)
)
