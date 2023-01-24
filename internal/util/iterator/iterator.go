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

// ErrIteratorDone is returned when the iterator is read to the end.
var ErrIteratorDone = errors.New("iterator is read to the end")

// Interface is an iterator interface.
type Interface[K, V any] interface {
	// Next returns the next key/value pair, where the key is a slice index, map key, document number, etc,
	// and the value is the slice or map value, next document, etc.
	// Returned error could be (possibly wrapped) ErrIteratorDone or some fatal error.
	Next() (K, V, error)

	// Close indicates that the iterator will no longer be used.
	// If Close is called, future calls to Next might panic.
	// If Close is not called, the iterator might leak resources or panic.
	// Close must be concurrency-safe and may be called multiple times.
	// All calls after the first should have no effect.
	Close()
}
