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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Array represents BSON array: an ordered collection of BSON values, accessed by 0-based indexes.
//
// Zero value is a valid empty array.
type Array struct {
	s      []any
	frozen bool
}

// MakeArray creates an empty array with set capacity.
func MakeArray(capacity int) *Array {
	if capacity == 0 {
		return new(Array)
	}

	return &Array{s: make([]any, 0, capacity)}
}

// NewArray creates an array with the given values.
//
// It panics if any of the values is not a valid BSON type.
func NewArray(values ...any) (*Array, error) {
	for _, v := range values {
		assertType(v)
	}

	return &Array{s: values}, nil
}

func (a *Array) compositeType() {}

// Freeze prevents array from further modifications.
// Any methods that would modify the array will panic.
//
// It is safe to call Freeze multiple times.
func (a *Array) Freeze() {
	if a != nil {
		a.frozen = true
	}
}

// checkFrozen panics if array is frozen.
func (a *Array) checkFrozen() {
	if a.frozen {
		panic("array is frozen and can't be modified")
	}
}

// DeepCopy returns an unfrozen deep copy of this Array.
func (a *Array) DeepCopy() *Array {
	if a == nil {
		panic("types.Array.DeepCopy: nil array")
	}

	return deepCopy(a).(*Array)
}

// Len returns the number of values in the array.
//
// It returns 0 for nil Array.
func (a *Array) Len() int {
	if a == nil {
		return 0
	}

	return len(a.s)
}

// Iterator returns an iterator for the array.
func (a *Array) Iterator() iterator.Interface[int, any] {
	return newArrayIterator(a)
}

// Get returns a value at the given index.
func (a *Array) Get(index int) (any, error) {
	if l := a.Len(); index < 0 || index >= l {
		return nil, fmt.Errorf("types.Array.Get: index %d is out of bounds [0-%d)", index, l)
	}

	return a.s[index], nil
}

// GetByPath returns a value by path.
func (a *Array) GetByPath(path Path) (any, error) {
	return getByPath(a, path)
}

// Set sets the value at the given index.
//
// It panics if the value is not a valid BSON type.
func (a *Array) Set(index int, value any) error {
	assertType(value)
	a.checkFrozen()

	if l := a.Len(); index < 0 || index >= l {
		return fmt.Errorf("types.Array.Set: index %d is out of bounds [0-%d)", index, l)
	}

	a.s[index] = value
	return nil
}

// Append appends given values to the array.
//
// It panics if any of the values is not a valid BSON type.
func (a *Array) Append(values ...any) {
	for _, v := range values {
		assertType(v)
	}

	a.checkFrozen()

	if a.s == nil {
		a.s = values
		return
	}

	a.s = append(a.s, values...)
}

// RemoveByPath removes (cuts) value by path, doing nothing if path points to nothing.
func (a *Array) RemoveByPath(path Path) {
	a.checkFrozen()

	removeByPath(a, path)
}

// Min returns the minimum value from the array.
func (a *Array) Min() any {
	if a == nil || a.Len() == 0 {
		panic("cannot get Min value; array is nil or empty")
	}

	min := must.NotFail(a.Get(0))
	for i := 1; i < a.Len(); i++ {
		value := must.NotFail(a.Get(i))
		if CompareOrder(min, value, Ascending) == Greater {
			min = value
		}
	}

	return min
}

// Max returns the maximum value from the array.
func (a *Array) Max() any {
	if a == nil || a.Len() == 0 {
		panic("cannot get Max value; array is nil or empty")
	}

	max := must.NotFail(a.Get(0))
	for i := 1; i < a.Len(); i++ {
		value := must.NotFail(a.Get(i))
		if CompareOrder(max, value, Ascending) == Less {
			max = value
		}
	}

	return max
}

// FilterArrayByType returns a new array which contains
// only values of the same BSON type as ref.
// All numbers are treated as the same type.
func (a *Array) FilterArrayByType(ref any) *Array {
	refType := detectDataType(ref)
	arr := MakeArray(0)

	for i := 0; i < a.Len(); i++ {
		value := must.NotFail(a.Get(i))
		vType := detectDataType(value)

		if refType == vType {
			arr.Append(value)
		}
	}

	return arr
}

// Contains checks if the Array contains the given value.
//
// It panics if the filterValue is not a valid BSON type.
func (a *Array) Contains(filterValue any) bool {
	assertType(filterValue)

	switch filterValue := filterValue.(type) {
	case *Document, *Array:
		// filterValue is a composite type, so either a and filterValue must be equal
		// or at least one element of a must be equal with filterValue.

		if res := Compare(a, filterValue); res == Equal {
			return true
		}

		for _, elem := range a.s {
			if res := Compare(elem, filterValue); res == Equal {
				return true
			}
		}
		return false

	default:
		// filterValue is a scalar, so we compare it to each scalar element of a
		for _, elem := range a.s {
			switch elem := elem.(type) {
			case *Document, *Array:
				// we need elem and filterValue to be exactly equal, so we do nothing here
			default:
				if compareScalars(elem, filterValue) == Equal {
					return true
				}
			}
		}
		return false
	}
}

// ContainsAll checks if Array a contains all the given values of Array b.
// Currently, this algorithm is O(n^2) without any performance tuning.
// This place can be significantly improved if a more performant algorithm is chosen.
func (a *Array) ContainsAll(b *Array) bool {
	for _, v := range b.s {
		if !a.Contains(v) {
			return false
		}
	}
	return true
}

// Remove removes the value at the given index.
func (a *Array) Remove(index int) {
	a.checkFrozen()

	if l := a.Len(); index < 0 || index >= l {
		panic("types.Array.Remove: index is out of bounds")
	}

	a.s = append(a.s[:index], a.s[index+1:]...)
}
