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
	"fmt"
)

// RawDocument represents a BSON document a.k.a object in the binary encoded form.
type RawDocument []byte

// field represents a single Document field in the (partially) decoded form.
type field struct {
	value any
	name  string
}

// Document represents a BSON document a.k.a object in the (partially) decoded form.
//
// It may contain duplicate field names.
type Document struct {
	fields []field
}

func NewDocument(pairs ...any) (*Document, error) {
	l := len(pairs)
	if l%2 != 0 {
		return nil, fmt.Errorf("invalid number of arguments: %d", l)
	}

	res := &Document{
		fields: make([]field, l/2),
	}

	for i := 0; i < l; i += 2 {
		name, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("invalid field name type: %T", pairs[i])
		}

		value := pairs[i+1]
		if !validType(value) {
			return nil, fmt.Errorf("invalid field value type: %T", value)
		}

		res.fields[i/2] = field{
			name:  name,
			value: value,
		}
	}

	return res, nil
}

func (o *Document) All(yield func(name string, value any) bool) bool {
	for _, f := range o.fields {
		if !yield(f.name, f.value) {
			return false
		}
	}

	return true
}
