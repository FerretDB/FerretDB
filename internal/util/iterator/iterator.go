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

// Package iterator describes a generic Iterator interface and related utilities.
package iterator

import "errors"

// ErrIteratorDone is returned when the iterator is read to the end or closed.
var ErrIteratorDone = errors.New("iterator is read to the end or closed")

// Interface is an iterator interface.
type Interface[K, V any] interface {
	// Next returns the next key/value pair, where the key is a slice index, map key, document number, etc,
	// and the value is the slice or map value, next document, etc.
	//
	// Returned error could be (possibly wrapped) ErrIteratorDone or some fatal error
	// like (possibly wrapped) context.Canceled.
	// In any case, even if iterator was read to the end, and Next returned ErrIteratorDone,
	// or Next returned fatal error,
	// Close method still should be called.
	//
	// Next should not be called concurrently with other Next calls,
	// but it can be called concurrently with Close.
	Next() (K, V, error)

	// Close indicates that the iterator will no longer be used.
	// After Close is called, future calls to Next must return ErrIteratorDone,
	// even if previous call returned a different error.
	//
	// Close must be called.
	// If it wasn't, the iterator might leak resources or panic later.
	//
	// Close must be concurrency-safe and may be called multiple times.
	// All calls after the first should have no observable effect.
	Close()
}
