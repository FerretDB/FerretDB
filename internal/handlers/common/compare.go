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
		if filter, ok := filter.(*types.Document); ok {
			switch compareDocuments(filter, docValue) {
			case equal:
				return equal
			case greater:
				return greater
			case less:
				return less
			}
		}
		return notEqual

	case *types.Array:
		if filter, ok := filter.(*types.Array); ok {
			switch compareArrays(filter, docValue) {
			case equal:
				return equal
			case greater:
				return greater
			case less:
				return less
			}
			return notEqual
		}

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

	if !slices.Equal(a.Keys(), b.Keys()) {
		return false
	}
	return reflect.DeepEqual(a.Map(), b.Map())
}

// compareDocuments compares 2 documents.
func compareDocuments(f, d *types.Document) compareResult {
	if f == nil {
		log.Panicf("%v is nil", f)
	}
	if d == nil {
		log.Panicf("%v is nil", d)
	}

	if !slices.Equal(f.Keys(), d.Keys()) {
		return notEqual
	}

	compareResult := notEqual
	for kf, vf := range f.Map() {
		vd := d.Map()[kf]
		res := compare(vd, vf)
		if res == notEqual {
			return notEqual
		}
		if res == equal {
			continue
		}
		if compareResult == notEqual {
			compareResult = res
		}
		if compareResult != res {
			return notEqual
		}
	}

	return compareResult
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

// compareArrays compares indices of the filter array according to
// indices of the document array and return documents which array is greater or lesser
func compareArrays(filterArr, docArr *types.Array) compareResult {
	if docArr.Len() == 0 && filterArr.Len() == 0 {
		return equal
	}

	compareResult := notEqual
	for i := 0; i < docArr.Len(); i++ {
		arrValue := must.NotFail(docArr.Get(i))
		switch docArrValue := arrValue.(type) {
		case *types.Array:
			if filterArrValue, ok := must.NotFail(filterArr.Get(0)).(*types.Array); ok {
				for i := 0; i < docArrValue.Len(); i++ {
					arrSubValue := must.NotFail(docArrValue.Get(i))
					if _, ok := arrSubValue.(*types.Array); ok {
						continue
					}

					filterArrValue := must.NotFail(filterArrValue.Get(i))
					res := compare(arrSubValue, filterArrValue)
					if res == notEqual {
						return notEqual
					}

					// define which way filter array is going to be (greater or lesser)
					if compareResult == notEqual {
						compareResult = res
					}
					// if a conflict occurs on subsequent iterations, for example,
					// one value is greater and the other is lesser, return not eq
					if compareResult != res {
						return notEqual
					}
				}
			} else {
				return greater
			}
		case *types.Document:
			if filterValue, ok := must.NotFail(filterArr.Get(i)).(*types.Document); ok {
				res := compareDocuments(filterValue, docArrValue)
				if res == notEqual {
					return notEqual
				}
				if compareResult == notEqual {
					compareResult = res
				}
				if compareResult != res {
					return notEqual
				}
			} else {
				return greater
			}

		default:
			// check the last element, if it is not an array return greater, bcs
			// there are not enough elements in the filter array
			if i >= filterArr.Len() && docArr.Len()-1 == i {
				return greater
			}

			// skip element for possibly similar doc subarray
			if i >= filterArr.Len() {
				continue
			}

			filterValue := must.NotFail(filterArr.Get(i))
			res := compare(docArrValue, filterValue)
			if res == notEqual {
				return notEqual
			}
			if res == equal {
				continue
			}
			if compareResult == notEqual {
				compareResult = res
			}
			if compareResult != res {
				return notEqual
			}
		}
	}

	return compareResult
}
