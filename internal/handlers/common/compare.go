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
	"log"
	"math"
	"math/big"
	"reflect"
	"time"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
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
			if math.IsNaN(a) && math.IsNaN(b) {
				return equal
			}
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
		if !ok {
			return notEqual
		}
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

	case types.ObjectID:
		b, ok := b.(types.ObjectID)
		if !ok {
			return notEqual
		}
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
		b, ok := b.(bool)
		if !ok {
			return notEqual
		}
		if a == b {
			return equal
		}
		if b {
			return less
		}
		return greater

	case time.Time:
		b, ok := b.(time.Time)
		if !ok {
			return notEqual
		}
		return compareOrdered(a.UnixMilli(), b.UnixMilli())

	case types.NullType:
		_, ok := b.(types.NullType)
		if ok {
			return equal
		}
		return notEqual

	case types.Regex:
		b, ok := b.(types.Regex)
		if ok && a == b {
			return equal
		}
		return notEqual

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

	default:
		panic(fmt.Sprintf("unhandled type %T", a))
	}
}

// compare compares the filter to the value of the document, whether it is a composite type or a scalar type.
func compare(docValue, filter any) compareResult {
	if docValue == nil {
		panic("docValue is nil")
	}
	if filter == nil {
		panic("filter is nil")
	}

	switch docValue := docValue.(type) {
	case *types.Document:
		// TODO: implement document comparing
		return notEqual

	case *types.Array:
		for i := 0; i < docValue.Len(); i++ {
			arrValue := must.NotFail(docValue.Get(i))
			switch arrValue.(type) {
			case *types.Document, *types.Array:
				continue
			}

			switch compareScalars(arrValue, filter) {
			case equal:
				return equal
			case greater:
				return greater
			case less:
				return less
			case notEqual:
				continue
			}
		}
		return notEqual

	default:
		return compareScalars(docValue, filter)
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
func compareNumbers(a float64, b int64) compareResult {
	if math.IsNaN(a) {
		return notEqual
	}

	// TODO figure out correct precision
	bigFloat := new(big.Float).SetFloat64(a).SetPrec(100000)
	bigFloatFromInt := new(big.Float).SetInt64(b).SetPrec(100000)

	switch bigFloat.Cmp(bigFloatFromInt) {
	case -1:
		return less
	case 0:
		return equal
	case 1:
		return greater
	default:
		panic("not reached")
	}
}

// matchDocuments returns true if 2 documents are equal.
func matchDocuments(a, b *types.Document) bool {
	if a == nil {
		log.Panicf("%v is nil", a)
	}
	if b == nil {
		log.Panicf("%v is nil", b)
	}

	keys := a.Keys()
	if !slices.Equal(keys, b.Keys()) {
		return false
	}
	return reflect.DeepEqual(a.Map(), b.Map())
}

// matchArrays returns true if a filter array equals exactly the specified array or
// array contains an element that equals the array.
func matchArrays(filterArr, docArr *types.Array) bool {
	if filterArr == nil {
		log.Panicf("%v is nil", filterArr)
	}
	if docArr == nil {
		log.Panicf("%v is nil", docArr)
	}

	if string(must.NotFail(fjson.Marshal(filterArr))) == string(must.NotFail(fjson.Marshal(docArr))) {
		return true
	}

	for i := 0; i < docArr.Len(); i++ {
		arrValue := must.NotFail(docArr.Get(i))
		if arrValue, ok := arrValue.(*types.Array); ok {
			if string(must.NotFail(fjson.Marshal(filterArr))) == string(must.NotFail(fjson.Marshal(arrValue))) {
				return true
			}
		}
	}

	return false
}
