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

package common

import (
	"bytes"
	"fmt"
	"time"

	"golang.org/x/exp/constraints"

	"github.com/FerretDB/FerretDB/internal/types"
)

// compareResult represents the result of a comparison.
type compareResult int

const (
	equal compareResult = iota
	less
	greater
	notEqual // but not less or greater; for example, two NaNs
)

// compareScalars compares two scalar values.
func compareScalars(a, b any) compareResult {
	if a == nil {
		panic("a is nil")
	}
	if b == nil {
		panic("b is nil")
	}

	switch a := a.(type) {
	case float64:
		switch b := b.(type) {
		case float64:
			return compareOrdered(a, b)
		case int32:
			return compareNumbers(a, int64(b))
		case int64:
			return compareNumbers(a, b)
		default:
			return notEqual
		}

	case string:
		b, ok := b.(string)
		if ok {
			return compareOrdered(a, b)
		}
		return notEqual

	case types.Binary:
		b, ok := b.(types.Binary)
		if ok {
			al, bl := len(a.B), len(b.B)
			if al != bl {
				return compareOrdered(al, bl)
			}
			if a.Subtype != b.Subtype {
				return compareOrdered(a.Subtype, b.Subtype)
			}
			switch bytes.Compare(a.B, b.B) {
			case 0:
				return equal
			case -1:
				return less
			case 1:
				return greater
			default:
				panic("unreachable")
			}
		}
		return notEqual

	case types.ObjectID:
		b, ok := b.(types.ObjectID)
		if ok {
			switch bytes.Compare(a[:], b[:]) {
			case 0:
				return equal
			case -1:
				return less
			case 1:
				return greater
			default:
				panic("unreachable")
			}
		}
		return notEqual

	case bool:
		b, ok := b.(bool)
		if ok {
			if a == b {
				return equal
			}
			if b {
				return less
			}
			return greater
		}
		return notEqual

	case time.Time:
		b, ok := b.(time.Time)
		if ok {
			return compareOrdered(a.UnixNano(), b.UnixNano())
		}
		return notEqual

	case types.NullType:
		_, ok := b.(types.NullType)
		if ok {
			return equal
		}
		return notEqual

	case types.Regex:
		return notEqual // ???

	case int32:
		switch b := b.(type) {
		case float64:
			return filterCompareInvert(compareNumbers(b, int64(a)))
		case int32:
			return compareOrdered(a, b)
		case int64:
			return compareOrdered(int64(a), b)
		default:
			return notEqual
		}

	case types.Timestamp:
		b, ok := b.(types.Timestamp)
		if ok {
			return compareOrdered(a, b)
		}
		return notEqual

	case int64:
		switch b := b.(type) {
		case float64:
			return filterCompareInvert(compareNumbers(b, a))
		case int32:
			return compareOrdered(a, int64(b))
		case int64:
			return compareOrdered(a, b)
		default:
			return notEqual
		}

	case *types.Document:
		return notEqual

	case *types.Array:
		for i := 0; i < a.Len(); i++ {
			a, err := a.Get(i)
			if err != nil {
				panic(fmt.Sprintf("cannot get value from array, err is %v, array is %v, index is %v", err, a, i))
			}
			switch compareScalars(a, b) {
			case equal:
				return equal
			case greater:
				return greater
			case less:
				return less
			}
		}
		return notEqual

	default:
		panic(fmt.Sprintf("unhandled type %T", a))
	}
}

// filterCompareInvert swaps less and greater, keeping equal and notEqual.
func filterCompareInvert(res compareResult) compareResult {
	switch res {
	case equal:
		return equal
	case less:
		return greater
	case greater:
		return less
	case notEqual:
		return notEqual
	default:
		panic("unreachable")
	}
}

// compareOrdered compares two values of the same type using ==, <, > operators.
func compareOrdered[T constraints.Ordered](a, b T) compareResult {
	if a == b {
		return equal
	}
	if a < b {
		return less
	}
	if a > b {
		return greater
	}
	return notEqual
}

// compareNumbers compares two numbers.
//
// TODO https://github.com/FerretDB/FerretDB/issues/371
func compareNumbers(a float64, b int64) compareResult {
	return compareOrdered(a, float64(b))
}
