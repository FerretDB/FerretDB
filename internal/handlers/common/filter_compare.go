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

type compareResult int

const (
	equal compareResult = iota
	less
	greater
	notEqual // but not less or greater; for example, two NaNs
)

// filterCompareScalars returns true if given scalar values are equal as used by filters.
func filterCompareScalars(a, b any) compareResult {
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
			return filterCompareOrdered(a, b)
		case int32:
			return filterCompareNumbers(a, int64(b))
		case int64:
			return filterCompareNumbers(a, b)
		default:
			panic(fmt.Sprintf("unexpected type %T", b))
		}

	case string:
		b := b.(string)
		return filterCompareOrdered(a, b)

	case types.Binary:
		b := b.(types.Binary)
		al, bl := len(a.B), len(b.B)
		if al != bl {
			return filterCompareOrdered(al, bl)
		}
		if a.Subtype != b.Subtype {
			return filterCompareOrdered(a.Subtype, b.Subtype)
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

	case types.ObjectID:
		b := b.(types.ObjectID)
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

	case bool:
		b := b.(bool)
		if a == b {
			return equal
		}
		if b {
			return less
		}
		return greater

	case time.Time:
		b := b.(time.Time)
		return filterCompareOrdered(a.UnixNano(), b.UnixNano())

	case types.NullType:
		_ = b.(types.NullType)
		return equal // or notEqual?

	case types.Regex:
		_ = b.(types.Regex)
		return notEqual // ???

	case int32:
		switch b := b.(type) {
		case float64:
			return filterCompareInvert(filterCompareNumbers(b, int64(a)))
		case int32:
			return filterCompareOrdered(a, b)
		case int64:
			return filterCompareOrdered(int64(a), b)
		default:
			panic(fmt.Sprintf("unexpected type %T", b))
		}

	case types.Timestamp:
		b := b.(types.Timestamp)
		return filterCompareOrdered(a, b)

	case int64:
		switch b := b.(type) {
		case float64:
			return filterCompareInvert(filterCompareNumbers(b, a))
		case int32:
			return filterCompareOrdered(a, int64(b))
		case int64:
			return filterCompareOrdered(a, b)
		default:
			panic(fmt.Sprintf("unexpected type %T", b))
		}

	default:
		panic(fmt.Sprintf("unhandled type %T", a))
	}
}

// filterCompareInvert swaps less and greater, keeping equal and notEqual.
func filterCompareInvert(res compareResult) compareResult {
	switch res {
	case less:
		return greater
	case greater:
		return less
	default:
		return res
	}
}

// filterCompareOrdered compares two values of the same type using ==, <, > operators.
func filterCompareOrdered[T constraints.Ordered](a, b T) compareResult {
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

// filterCompareNumbers compares two numbers.
//
// https://github.com/FerretDB/FerretDB/issues/371
func filterCompareNumbers(a float64, b int64) compareResult {
	return filterCompareOrdered(a, float64(b))
}
