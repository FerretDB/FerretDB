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
	"strconv"
)

// getByPath returns Array or Object value by path - a sequence of indexes and keys.
func getByPath(str any, path ...string) (any, error) {
	for _, p := range path {
		switch s := str.(type) {
		case Array:
			index, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("types.getByPath: %w", err)
			}
			str, err = s.Get(index)
			if err != nil {
				return nil, fmt.Errorf("types.getByPath: %w", err)
			}

		case Document:
			var err error
			str, err = s.Get(p)
			if err != nil {
				return nil, fmt.Errorf("types.getByPath: %w", err)
			}

		default:
			return nil, fmt.Errorf("types.getByPath: can't access %T by path %q", str, p)
		}
	}

	return str, nil
}
