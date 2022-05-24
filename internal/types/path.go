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

// Path represents field path type. It should be used wherever we work with paths or dot notation.
type Path struct {
	s []string
}

// Len returns path length.
func (p Path) Len() int {
	return len(p.s)
}

// Slice returns path slice copy.
func (p Path) Slice() []string {
	path := make([]string, len(p.s))
	copy(path, p.s)
	return path
}

// Element returns path element specified by the index.
func (p Path) Element(index int) string {
	if len(p.s) < index || index < 0 {
		return ""
	}
	return p.s[index]
}

// RemoveFirstElement returns new path without first path element.
func (p Path) RemoveFirstElement() Path {
	if len(p.s) < 1 {
		return p
	}
	path := Path{s: make([]string, len(p.s)-1)}
	copy(path.s, p.s[1:])
	return path
}

// NewPath returns Path from strings slice.
func NewPath(path ...string) Path {
	p := Path{s: make([]string, len(path))}
	copy(p.s, path)
	return p
}

// NewPathFromString returns Path from path string.
func NewPathFromString(s string) Path {
	path := strings.Split(s, ".")

	return NewPath(path...)
}

// RemoveByPath removes document by path, doing nothing if the key does not exist.
func RemoveByPath[T CompositeTypeInterface](comp T, path Path) {
	removeByPath(comp, path)
}

// getByPath returns a value by path - a sequence of indexes and keys.
func getByPath[T CompositeTypeInterface](comp T, path Path) (any, error) {
	var next any = comp
	for _, p := range path.Slice() {
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

// removeByPath removes path elements for given value, which could be *Document or *Array.
func removeByPath(v any, path Path) {
	if path.Len() == 0 {
		return
	}

	key := path.Element(0)
	switch v := v.(type) {
	case *Document:
		if _, ok := v.m[key]; !ok {
			return
		}
		if path.Len() == 1 {
			v.Remove(key)
			return
		}
		removeByPath(v.m[key], path.RemoveFirstElement())

	case *Array:
		i, err := strconv.Atoi(key)
		if err != nil {
			return // no such path
		}
		if i > len(v.s)-1 {
			return // no such path
		}
		if path.Len() == 1 {
			v.s = append(v.s[:i], v.s[i+1:]...)
			return
		}
		removeByPath(v.s[i], path.RemoveFirstElement())
	default:
		// no such path: scalar value
	}
}
