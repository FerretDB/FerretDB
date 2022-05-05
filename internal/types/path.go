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

// RemoveByPath removes document by path, doing nothing if the key does not exist.
func RemoveByPath[T CompositeTypeInterface](comp T, keys ...string) {
	removeByPath(comp, keys...)
}

// getByPath returns a value by path - a sequence of indexes and keys.
func getByPath[T CompositeTypeInterface](comp T, path ...string) (any, error) {
	var next any = comp
	for _, p := range path {
		switch s := next.(type) {
		case *Document:
			var err error
			next, err = s.Get(p)
			if err != nil {
				return nil, fmt.Errorf("types.getByPath: %w", err)
			}

		case *Array:
			index, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("types.getByPath: %w", err)
			}
			next, err = s.Get(index)
			if err != nil {
				return nil, fmt.Errorf("types.getByPath: %w", err)
			}

		default:
			return nil, fmt.Errorf("types.getByPath: can't access %T by path %q", next, p)
		}
	}

	return next, nil
}

func removeByPath(v any, keys ...string) {
	if len(keys) == 0 {
		return
	}

	key := keys[0]
	switch v := v.(type) {
	case *Document:
		if _, ok := v.m[key]; !ok {
			return
		}
		if len(keys) == 1 {
			v.Remove(key)
			return
		}
		removeByPath(v.m[key], keys[1:]...)

	case *Array:
		i, err := strconv.Atoi(key)
		if err != nil {
			return // no such path
		}
		if i > len(v.s)-1 {
			return // no such path
		}
		if len(keys) == 1 {
			v.s = append(v.s[:i], v.s[i+1:]...)
			return
		}
		removeByPath(v.s[i], keys[1:]...)
	default:
		// no such path: scalar value
	}
}
