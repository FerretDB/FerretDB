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

package iterator

// Closer is an interface for closing iterators.
type Closer interface {
	Close()
}

// MultiCloser is a helper for closing multiple closers.
type MultiCloser struct {
	closers []Closer
}

// NewMultiCloser returns a new MultiCloser for non-nil closers.
func NewMultiCloser(closers ...Closer) *MultiCloser {
	mc := new(MultiCloser)
	mc.Add(closers...)

	return mc
}

// Add adds non-nil closers to the MultiCloser.
func (mc *MultiCloser) Add(closers ...Closer) {
	for _, c := range closers {
		if c != nil {
			mc.closers = append(mc.closers, c)
		}
	}
}

// Close closes all added closers.
func (mc *MultiCloser) Close() {
	for _, c := range mc.closers {
		c.Close()
	}
}
