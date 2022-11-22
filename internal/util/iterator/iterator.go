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

// Package iterator describes a generic Iterator interface.
package iterator

import "errors"

// ErrIteratorDone  is returned when the iterator is read to the end.
var ErrIteratorDone = errors.New("iterator is read to the end")

// Interface is an iterator interface.
type Interface[E1, E2 any] interface {
	// Next returns an ordered pair.
	// The meaning of that ordered pair depends on the type of the iterator.
	// For example, for maps E1 is key and E2 is value, for slices E1 is index and E2 is value.
	// If the iterator is at the end, it returns possibly wrapped ErrEndOfIterator as error.
	// Other errors may be returned as well, they depend on the implementation and could be wrapped.
	// Next returns the next key/value pair, where key is slice index, map key, document number, etc, and the value is the slice or map value, next document, etc.
	// Returned error could be (possibly wrapped) ErrEndOfIterator or some fatal error.
	Next() (E1, E2, error)
}
