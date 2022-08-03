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

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Path represents the field path type. It should be used wherever we work with paths or dot notation.
// Path should be stored and passed as a value. Its methods return new values, not modifying the receiver's state.
type Path struct {
	s []string
}

// NewPath returns Path from a strings slice.
func NewPath(path []string) Path {
	if len(path) == 0 {
		panic("empty path")
	}
	for _, s := range path {
		if s == "" {
			panic("path element must not be empty")
		}
	}
	p := Path{s: make([]string, len(path))}
	copy(p.s, path)
	return p
}

// NewPathFromString returns Path from path string. Path string should contain fields separated with '.'.
func NewPathFromString(s string) Path {
	path := strings.Split(s, ".")

	return NewPath(path)
}

// String returns dot-separated path value.
func (p Path) String() string {
	return strings.Join(p.s, ".")
}

// Len returns path length.
func (p Path) Len() int {
	return len(p.s)
}

// Slice returns path values array.
func (p Path) Slice() []string {
	path := make([]string, p.Len())
	copy(path, p.s)
	return path
}

// Suffix returns the last path element.
func (p Path) Suffix() string {
	if len(p.s) <= 1 {
		panic("path should have more than 1 element")
	}
	return p.s[p.Len()-1]
}

// Prefix returns the first path element.
func (p Path) Prefix() string {
	if p.Len() <= 1 {
		panic("path should have more than 1 element")
	}
	return p.s[0]
}

// TrimSuffix returns a path without the last element.
func (p Path) TrimSuffix() Path {
	if p.Len() <= 1 {
		panic("path should have more than 1 element")
	}
	return NewPath(p.s[:p.Len()-1])
}

// TrimPrefix returns a copy of path without the first element.
func (p Path) TrimPrefix() Path {
	if p.Len() <= 1 {
		panic("path should have more than 1 element")
	}
	return NewPath(p.s[1:])
}

// Append returns new Path constructed from the current path and given element.
func (p Path) Append(elem string) Path {
	elems := p.Slice()

	elems = append(elems, elem)

	return NewPath(elems)
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

	var key string
	if path.Len() == 1 {
		key = path.String()
	} else {
		key = path.Prefix()
	}
	switch v := v.(type) {
	case *Document:
		if _, ok := v.m[key]; !ok {
			return
		}
		if path.Len() == 1 {
			v.Remove(key)
			return
		}
		removeByPath(v.m[key], path.TrimPrefix())

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
		removeByPath(v.s[i], path.TrimPrefix())
	default:
		// no such path: scalar value
	}
}

// insertByPath inserts missing parts of the path into Document.
func insertByPath(doc *Document, path Path) error {
	var next any = doc

	var insertedPath Path
	for _, pathElem := range path.TrimSuffix().Slice() {
		insertedPath = insertedPath.Append(pathElem)

		v, err := doc.GetByPath(insertedPath)
		if err != nil {
			suffix := len(insertedPath.Slice()) - 1
			if suffix < 0 {
				panic("invalid path")
			}

			switch v := next.(type) {
			case *Document:
				must.NoError(v.Set(insertedPath.Slice()[suffix], must.NotFail(NewDocument())))
			case *Array:
				_, err := strconv.Atoi(insertedPath.Slice()[suffix])
				if err != nil {
					return fmt.Errorf(
						"Cannot create field '%s' in element {%s: %s}",
						pathElem,
						insertedPath.Slice()[suffix-1],
						formatDocumentValue(v),
					)
				}
			}

			next = must.NotFail(doc.GetByPath(insertedPath)).(*Document)

			continue
		}

		next = v
	}

	return nil
}

// formatDocumentValue formats document value for error message output.
func formatDocumentValue(v any) string {
	switch v := v.(type) {
	case *Document:
		return formatDocument(v)
	case *Array:
		return formatArray(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatDocument formats Document for error output.
func formatDocument(doc *Document) string {
	result := "{"

	for i, key := range doc.keys {
		if i > 0 {
			result += ", "
		}

		switch value := doc.m[key].(type) {
		case *Document:
			result += fmt.Sprintf("%q: %s", key, value)
		case *Array:
			result += fmt.Sprintf("%q: %s", key, value)
		case string:
			result += fmt.Sprintf(`%q: "%q"`, key, value)
		case NullType:
			result += fmt.Sprintf("%q: null", key)
		case nil:
			result += fmt.Sprintf("%q: null", key)
		case int32, int64:
			result += fmt.Sprintf("%q: %d", key, value)
		default:
			result += fmt.Sprintf("%q: %s", key, doc.m[key])
		}
	}

	return result + "}"
}

// formatArray formats Array for error output.
func formatArray(array *Array) string {
	result := "[ "

	for _, elem := range array.s {
		switch elem := elem.(type) {
		case *Document:
			result += fmt.Sprintf("%s, ", elem)
		case *Array:
			result += fmt.Sprintf("%s, ", elem)
		case string:
			result += fmt.Sprintf(`"%s", `, elem)
		case NullType:
			result += "null, "
		case int32, int64:
			result += fmt.Sprintf("%d, ", elem)
		default:
			result += fmt.Sprintf("%s, ", elem)
		}
	}

	result = strings.TrimSuffix(result, ", ")

	return result + " ]"
}
