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

// Package xiter provides iterator utilities.
//
// It may or may not be superseded by `x/exp/xiter` package.
// See https://github.com/golang/go/issues/61898.
package xiter

import "iter"

// Chunk returns an iterator over consecutive slices of up to n elements of seq.
// All but the last slice will have size n.
// All slices are clipped to have no capacity beyond the length.
// If seq is empty, the sequence is empty: there is no empty slice in the sequence.
// Chunk panics if n is less than 1.
//
// https://github.com/golang/go/issues/61898#issuecomment-2522037782
func Chunk[E any](seq iter.Seq[E], n int) iter.Seq[[]E] {
	if n < 1 {
		panic("cannot be less than 1")
	}

	return func(yield func([]E) bool) {
		var batch []E

		for e := range seq {
			if batch == nil {
				batch = make([]E, 0, n)
			}

			batch = append(batch, e)

			if len(batch) == n {
				if !yield(batch) {
					return
				}

				batch = nil
			}
		}

		if l := len(batch); l > 0 {
			yield(batch[:l:l])
		}
	}
}
