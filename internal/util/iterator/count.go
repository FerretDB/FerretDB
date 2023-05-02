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

import "errors"

// ConsumeCount returns the number of elements in the iterator.
// ErrIteratorDone error is returned as nil; any other error is returned as-is.
//
// Iterator is always closed at the end.
func ConsumeCount[K, V any](iter Interface[K, V]) (int, error) {
	defer iter.Close()

	var count int
	var err error

	for {
		_, _, err = iter.Next()
		if err == nil {
			count++
			continue
		}

		if errors.Is(err, ErrIteratorDone) {
			err = nil
		}

		return count, err
	}
}
