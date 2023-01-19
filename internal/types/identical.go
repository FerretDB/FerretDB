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

package types

import (
	"bytes"
	"errors"
	"time"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// Identical returns true if a and b are the same type
// and has the same value.
func Identical(a, b any) bool {
	switch a := a.(type) {
	case *Document:
		b, ok := b.(*Document)
		if !ok {
			return false
		}

		if a.Len() != b.Len() {
			return false
		}

		aIter, bIter := a.Iterator(), b.Iterator()

		for {
			_, aField, err := aIter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				return true
			} else if err != nil {
				return false
			}

			_, bField, err := bIter.Next()
			if err != nil {
				return false
			}

			if !Identical(aField, bField) {
				return false
			}
		}
	case *Array:
		b, ok := b.(*Array)
		if !ok {
			return false
		}

		if a.Len() != b.Len() {
			return false
		}

		aIter, bIter := a.Iterator(), b.Iterator()

		for {
			_, aItem, err := aIter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				return true
			} else if err != nil {
				return false
			}

			_, bItem, err := bIter.Next()
			if err != nil {
				return false
			}

			if !Identical(aItem, bItem) {
				return false
			}
		}
	case float64:
		b, ok := b.(float64)
		if !ok {
			return false
		}

		return a == b
	case string:
		b, ok := b.(string)
		if !ok {
			return false
		}

		return a == b
	case Binary:
		b, ok := b.(Binary)
		if !ok {
			return false
		}

		if len(a.B) != len(b.B) {
			return false
		}

		if a.Subtype != b.Subtype {
			return false
		}

		return bytes.Equal(a.B, b.B)
	case ObjectID:
		b, ok := b.(ObjectID)
		if !ok {
			return false
		}

		return bytes.Equal(a[:], b[:])
	case bool:
		b, ok := b.(bool)
		if !ok {
			return false
		}

		return a == b
	case time.Time:
		b, ok := b.(time.Time)
		if !ok {
			return false
		}

		return a.UnixMilli() == b.UnixMilli()
	case NullType:
		_, ok := b.(NullType)
		return ok
	case Regex:
		b, ok := b.(Regex)
		if !ok {
			return false
		}

		return a == b
	case int32:
		b, ok := b.(int32)
		if !ok {
			return false
		}

		return a == b
	case Timestamp:
		b, ok := b.(Timestamp)
		if !ok {
			return false
		}

		return a == b
	case int64:
		b, ok := b.(int64)
		if !ok {
			return false
		}

		return a == b
	}

	panic("not reached")
}
