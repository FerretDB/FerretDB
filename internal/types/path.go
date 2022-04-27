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
	"strings"
)

// getPairByPath returns key/index and value pair by path - a sequence of indexes and keys separated by dots.
func getPairByPath[T CompositeTypeInterface](comp T, path string) (string, any, error) {
	var key string
	var val any = comp
	for _, key = range strings.Split(path, ".") {
		switch v := val.(type) {
		case *Document:
			var err error
			if val, err = v.Get(key); err != nil {
				return "", nil, fmt.Errorf("types.getPairByPath: %w", err)
			}

		case *Array:
			i, err := strconv.Atoi(key)
			if err != nil {
				return "", nil, fmt.Errorf("types.getPairByPath: %w", err)
			}
			if val, err = v.Get(i); err != nil {
				return "", nil, fmt.Errorf("types.getPairByPath: %w", err)
			}

		default:
			return "", nil, fmt.Errorf("types.getPairByPath: can't access %T by path %q", val, key)
		}
	}

	return key, val, nil
}
