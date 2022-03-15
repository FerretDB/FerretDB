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

// getKeyPaths returns a key path.
func getKeyPaths[T CompositeTypeInterface](comp T, key string, currentPath []string, in [][]string) (res [][]string, err error) {
	var next any = comp

	switch s := next.(type) {
	case *Document:
		for _, k := range s.Keys() {
			newPath := make([]string, len(currentPath)+1, len(currentPath)+1)
			copy(newPath, currentPath)
			newPath[len(newPath)-1] = k
			if k == key {
				res = append(res, newPath)
				continue
			}

			var p any

			p, err = s.Get(k)
			if err != nil {
				err = fmt.Errorf("types.getKeyPath: can't access %s", strings.Join(newPath, "."))
				return
			}

			switch t := p.(type) {
			case *Document:
				var deeper [][]string
				deeper, err = getKeyPaths(t, key, newPath, res)
				res = append(res, deeper...)
				if err != nil {
					return
				}

			case *Array:
				for i := 0; i < s.Len(); i++ {
					arrayPath := make([]string, 0, len(currentPath)+1)
					copy(arrayPath, newPath)
					arrayPath = append(arrayPath, strconv.Itoa(i))

					var deeper [][]string
					deeper, err = getKeyPaths(t, key, arrayPath, res)
					res = append(res, deeper...)
					if err != nil {
						return
					}
				}
			}

		}

	case *Array:
		for i := 0; i < s.Len(); i++ {
			newPath := make([]string, len(currentPath)+1, len(currentPath)+1)
			copy(newPath, currentPath)
			newPath[len(newPath)-1] = strconv.Itoa(i)
			var deeper [][]string
			deeper, err = getKeyPaths(comp, key, newPath, res)
			res = append(res, deeper...)
			if err != nil {
				return
			}
		}
	}

	return
}
