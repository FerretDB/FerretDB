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

import "fmt"

// Array represents BSON array.
//
// Zero value is a valid empty array.
type Array struct {
	s []any
}

// MakeArray creates an empty array with set capacity.
func MakeArray(capacity int) *Array {
	if capacity == 0 {
		return new(Array)
	}
	return &Array{s: make([]any, 0, capacity)}
}

// NewArray creates an array with the given values.
func NewArray(values ...any) (*Array, error) {
	for i, value := range values {
		if err := validateValue(value); err != nil {
			return nil, fmt.Errorf("types.NewArray: index %d: %w", i, err)
		}
	}

	return &Array{s: values}, nil
}

// MustNewArray is a NewArray that panics in case of error.
//
// Deprecated: use `must.NotFail(NewArray(...))` instead.
func MustNewArray(values ...any) *Array {
	a, err := NewArray(values...)
	if err != nil {
		panic(err)
	}
	return a
}

func (a *Array) compositeType() {}

// DeepCopy returns a deep copy of this Array.
func (a *Array) DeepCopy() *Array {
	if a == nil {
		panic("types.Array.DeepCopy: nil array")
	}
	return deepCopy(a).(*Array)
}

// Len returns the number of elements in the array.
//
// It returns 0 for nil Array.
func (a *Array) Len() int {
	if a == nil {
		return 0
	}
	return len(a.s)
}

// Get returns a value at the given index.
func (a *Array) Get(index int) (any, error) {
	if l := a.Len(); index < 0 || index >= l {
		return nil, fmt.Errorf("types.Array.Get: index %d is out of bounds [0-%d)", index, l)
	}

	return a.s[index], nil
}

// GetByPath returns a value by path - a sequence of indexes and keys.
func (a *Array) GetByPath(path ...string) (any, error) {
	return getByPath(a, path...)
}

// Set sets the value at the given index.
func (a *Array) Set(index int, value any) error {
	if l := a.Len(); index < 0 || index >= l {
		return fmt.Errorf("types.Array.Set: index %d is out of bounds [0-%d)", index, l)
	}

	if err := validateValue(value); err != nil {
		return fmt.Errorf("types.Array.Set: %w", err)
	}

	a.s[index] = value
	return nil
}

// Append appends given values to the array.
func (a *Array) Append(values ...any) error {
	for _, value := range values {
		if err := validateValue(value); err != nil {
			return fmt.Errorf("types.Array.Append: %w", err)
		}
	}

	if a.s == nil {
		a.s = values
		return nil
	}

	a.s = append(a.s, values...)
	return nil
}

// RemoveByPath removes document by path, doing nothing if the key does not exist.
func (a *Array) RemoveByPath(keys ...string) {
	removeByPath(a, keys...)
}
