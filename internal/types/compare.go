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
	"fmt"
	"math"
	"math/big"
	"time"

	"golang.org/x/exp/constraints"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../bin/stringer -linecomment -type CompareResult

// CompareResult represents the result of a comparison.
type CompareResult int8

const (
	Equal    CompareResult = 0   // ==
	Less     CompareResult = -1  // <
	Greater  CompareResult = 1   // >
	NotEqual CompareResult = 127 // !=
)

// Compare compares BSON values.
func Compare(a, b any) CompareResult {
	if a == nil {
		panic("a is nil")
	}
	if b == nil {
		panic("b is nil")
	}

	switch a := a.(type) {
	case *Document:
		// TODO: implement document comparing
		return NotEqual

	case *Array:
		for i := 0; i < a.Len(); i++ {
			v := must.NotFail(a.Get(i))
			switch v.(type) {
			case *Document, *Array:
				continue
			}

			if res := compareScalars(v, b); res != NotEqual {
				return res
			}
		}
		return NotEqual

	default:
		return compareScalars(a, b)
	}
}

// compareScalars compares BSON scalar values.
func compareScalars(a, b any) CompareResult {
	compareEnsureScalar(a)
	compareEnsureScalar(b)

	switch a := a.(type) {
	case float64:
		switch b := b.(type) {
		case float64:
			if math.IsNaN(a) && math.IsNaN(b) {
				return Equal
			}
			return compareOrdered(a, b)
		case int32:
			return compareNumbers(a, int64(b))
		case int64:
			return compareNumbers(a, b)
		default:
			return NotEqual
		}

	case string:
		b, ok := b.(string)
		if ok {
			return compareOrdered(a, b)
		}
		return NotEqual

	case Binary:
		b, ok := b.(Binary)
		if !ok {
			return NotEqual
		}
		al, bl := len(a.B), len(b.B)
		if al != bl {
			return compareOrdered(al, bl)
		}
		if a.Subtype != b.Subtype {
			return compareOrdered(a.Subtype, b.Subtype)
		}
		return CompareResult(bytes.Compare(a.B, b.B))

	case ObjectID:
		b, ok := b.(ObjectID)
		if !ok {
			return NotEqual
		}
		return CompareResult(bytes.Compare(a[:], b[:]))

	case bool:
		b, ok := b.(bool)
		if !ok {
			return NotEqual
		}
		if a == b {
			return Equal
		}
		if b {
			return Less
		}
		return Greater

	case time.Time:
		b, ok := b.(time.Time)
		if !ok {
			return NotEqual
		}
		return compareOrdered(a.UnixMilli(), b.UnixMilli())

	case NullType:
		_, ok := b.(NullType)
		if ok {
			return Equal
		}
		return NotEqual

	case Regex:
		b, ok := b.(Regex)
		if ok && a == b {
			return Equal
		}
		return NotEqual

	case int32:
		switch b := b.(type) {
		case float64:
			return compareInvert(compareNumbers(b, int64(a)))
		case int32:
			return compareOrdered(a, b)
		case int64:
			return compareOrdered(int64(a), b)
		default:
			return NotEqual
		}

	case Timestamp:
		b, ok := b.(Timestamp)
		if ok {
			return compareOrdered(a, b)
		}
		return NotEqual

	case int64:
		switch b := b.(type) {
		case float64:
			return compareInvert(compareNumbers(b, a))
		case int32:
			return compareOrdered(a, int64(b))
		case int64:
			return compareOrdered(a, b)
		default:
			return NotEqual
		}
	}

	panic("not reached")
}

// compareEnsureScalar panics if v is not a BSON scalar value.
func compareEnsureScalar(v any) {
	if v == nil {
		panic("v is nil")
	}

	switch v.(type) {
	case float64, string, Binary, ObjectID, bool, time.Time, NullType, Regex, int32, Timestamp, int64:
		return
	}

	panic(fmt.Sprintf("unhandled type %T", v))
}

// compareInvert swaps Less and Greater, keeping Equal and NotEqual.
func compareInvert(res CompareResult) CompareResult {
	switch res {
	case Equal:
		return Equal
	case Less:
		return Greater
	case Greater:
		return Less
	case NotEqual:
		return NotEqual
	}

	panic("not reached")
}

// compareOrdered compares BSON values of the same type using ==, <, > operators.
func compareOrdered[T constraints.Ordered](a, b T) CompareResult {
	switch {
	case a == b:
		return Equal
	case a < b:
		return Less
	case a > b:
		return Greater
	default:
		return NotEqual
	}
}

// compareNumbers compares BSON numbers.
func compareNumbers(a float64, b int64) CompareResult {
	if math.IsNaN(a) {
		return NotEqual
	}

	// TODO figure out correct precision
	bigA := new(big.Float).SetFloat64(a).SetPrec(100000)
	bigB := new(big.Float).SetInt64(b).SetPrec(100000)

	return CompareResult(bigA.Cmp(bigB))
}
