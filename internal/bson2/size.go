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

package bson2

import (
	"strconv"

	"github.com/cristalhq/bson/bsonproto"
)

// sizeAny returns a size of the encoding of value v in bytes.
//
// It panics for invalid types.
func sizeAny(v any) int {
	switch v := v.(type) {
	case *Document:
		return sizeDocument(v)
	case RawDocument:
		return len(v)
	case *Array:
		return sizeArray(v)
	case RawArray:
		return len(v)
	default:
		return bsonproto.SizeAny(v)
	}
}

// sizeDocument returns a size of the encoding of Document doc in bytes.
func sizeDocument(doc *Document) int {
	size := 5

	for _, f := range doc.fields {
		size += 1 + len(f.name) + 1 + sizeAny(f.value)
	}

	return size
}

// sizeArray returns a size of the encoding of Array arr in bytes.
func sizeArray(arr *Array) int {
	size := 5

	for i, v := range arr.elements {
		size += 1 + len(strconv.Itoa(i)) + 1 + sizeAny(v)
	}

	return size
}
