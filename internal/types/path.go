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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../bin/stringer -linecomment -type PathErrorCode

// PathErrorCode represents PathError code.
type PathErrorCode int

const (
	_ PathErrorCode = iota

	// ErrPathElementEmpty indicates that provided path contains an empty element.
	ErrPathElementEmpty

	// ErrPathElementInvalid indicates that provided path contains an invalid element (other than empty).
	ErrPathElementInvalid

	// ErrPathKeyNotFound indicates that key was not found in document.
	ErrPathKeyNotFound

	// ErrPathIndexInvalid indicates that provided array index is invalid.
	ErrPathIndexInvalid

	// ErrPathIndexOutOfBound indicates that provided array index is out of bound.
	ErrPathIndexOutOfBound

	// ErrPathCannotAccess indicates that path couldn't be accessed.
	ErrPathCannotAccess

	// ErrPathCannotCreateField indicates that it's impossible to create a specific field.
	ErrPathCannotCreateField

	// ErrPathConflictOverwrite indicates a path overwrites another path.
	ErrPathConflictOverwrite

	// ErrPathConflictCollision indicates a path creates collision at another path.
	ErrPathConflictCollision
)

// PathError describes an error that could occur on path related operations.
type PathError struct {
	err  error
	code PathErrorCode
}

// Error implements the error interface.
func (e *PathError) Error() string {
	return e.err.Error()
}

// Code returns the PathError code.
func (e *PathError) Code() PathErrorCode {
	return e.code
}

// newPathError creates a new PathError.
func newPathError(code PathErrorCode, reason error) error {
	return &PathError{err: reason, code: code}
}

// Path represents a parsed dot notation - a sequence of elements (document keys and array indexes) separated by dots.
//
// Path's elements can't be empty, include dots, spaces, or start with $.
//
// Path should be stored and passed as a value.
// Its methods return new values, not modifying the receiver's state.
type Path struct {
	e []string
}

// newPath returns Path from a strings slice.
func newPath(path ...string) (Path, error) {
	var res Path

	for _, e := range path {
		switch {
		case e == "":
			return res, newPathError(ErrPathElementEmpty, errors.New("path element must not be empty"))
		case strings.TrimSpace(e) != e:
			return res, newPathError(ErrPathElementInvalid, errors.New("path element must not contain spaces"))
		case strings.Contains(e, "."):
			return res, newPathError(ErrPathElementInvalid, errors.New("path element must contain '.'"))
			// TODO https://github.com/FerretDB/FerretDB/issues/3127
			// enable validation of `$` prefix
			// case strings.HasPrefix(e, "$"):
			//	return res, newPathError(ErrPathElementInvalid, errors.New("path element must not start with '$'"))
		}
	}

	res = Path{e: make([]string, len(path))}
	copy(res.e, path)

	return res, nil
}

// NewStaticPath returns Path from a strings slice.
//
// It panics on invalid paths. For that reason, it should not be used with user-provided paths.
func NewStaticPath(path ...string) Path {
	return must.NotFail(newPath(path...))
}

// NewPathFromString returns Path from a given dot notation.
//
// It returns an error if the path is invalid.
func NewPathFromString(s string) (Path, error) {
	return newPath(strings.Split(s, ".")...)
}

// String returns a dot notation for that path.
func (p Path) String() string {
	return strings.Join(p.e, ".")
}

// Len returns path length.
func (p Path) Len() int {
	return len(p.e)
}

// Slice returns path elements array.
func (p Path) Slice() []string {
	path := make([]string, p.Len())
	copy(path, p.e)
	return path
}

// Suffix returns the last path element.
func (p Path) Suffix() string {
	return p.e[p.Len()-1]
}

// Prefix returns the first path element.
func (p Path) Prefix() string {
	return p.e[0]
}

// TrimSuffix returns a path without the last element.
func (p Path) TrimSuffix() Path {
	if p.Len() <= 1 {
		panic("path should have more than 1 element")
	}

	return NewStaticPath(p.e[:p.Len()-1]...)
}

// TrimPrefix returns a copy of path without the first element.
func (p Path) TrimPrefix() Path {
	if p.Len() <= 1 {
		panic("path should have more than 1 element")
	}

	return NewStaticPath(p.e[1:]...)
}

// Append returns new Path constructed from the current path and given element.
func (p Path) Append(elem string) Path {
	elems := p.Slice()

	elems = append(elems, elem)

	return NewStaticPath(elems...)
}

// RemoveByPath removes document by path, doing nothing if the key does not exist.
func RemoveByPath[T CompositeTypeInterface](comp T, path Path) {
	removeByPath(comp, path)
}

// IsConflictPath returns PathError error if adding a path creates conflict at any of paths.
// Returned PathError error codes:
//
//   - ErrPathConflictOverwrite when path overwrites any paths: paths = []{{"a","b"}} path = {"a"};
//   - ErrPathConflictCollision when path creates collision:    paths = []{{"a"}}     path = {"a","b"};
func IsConflictPath(paths []Path, path Path) error {
	for _, p := range paths {
		target, prefix := p.Slice(), path.Slice()

		if len(target) < len(prefix) {
			target, prefix = prefix, target
		}

		if len(prefix) == 0 {
			panic("path cannot be empty string")
		}

		var different bool

		for i := range prefix {
			if prefix[i] != target[i] {
				different = true
				break
			}
		}

		if different {
			continue
		}

		if p.Len() >= path.Len() {
			return newPathError(ErrPathConflictOverwrite, errors.New("path overwrites previous path"))
		}

		// collisionPart is part of the path which creates collision, used in command error message.
		// If visitedPath is `a.b` and path is `a.b.c`, collisionPart is `b.c`.
		collisionPart := strings.Join(target[len(prefix):], ".")

		return newPathError(ErrPathConflictCollision, errors.New(collisionPart))
	}

	return nil
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
				return nil, newPathError(ErrPathKeyNotFound, fmt.Errorf("types.getByPath: %w", err))
			}

		case *Array:
			index, err := strconv.Atoi(p)
			if err != nil {
				return nil, newPathError(ErrPathIndexInvalid, fmt.Errorf("types.getByPath: %w", err))
			}

			if index < 0 {
				return nil, newPathError(
					ErrPathIndexInvalid,
					fmt.Errorf("types.getByPath: array index below zero: %d", index),
				)
			}

			next, err = s.Get(index)
			if err != nil {
				return nil, newPathError(ErrPathIndexOutOfBound, fmt.Errorf("types.getByPath: %w", err))
			}

		default:
			return nil, newPathError(
				ErrPathCannotAccess,
				fmt.Errorf("types.getByPath: can't access %T by path %q", next, p),
			)
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
		index := slices.Index(v.Keys(), key)
		if index == -1 {
			return
		}

		if path.Len() == 1 {
			v.Remove(key)
			return
		}

		removeByPath(v.fields[index].value, path.TrimPrefix())

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
				v.Set(insertedPath.Slice()[suffix], must.NotFail(NewDocument()))

			case *Array:
				ind, err := strconv.Atoi(insertedPath.Slice()[suffix])
				if err != nil {
					return newPathError(
						ErrPathCannotCreateField,
						fmt.Errorf(
							"Cannot create field '%s' in element {%s: %s}",
							pathElem,
							insertedPath.Slice()[suffix-1],
							FormatAnyValue(v),
						),
					)
				}

				if ind < 0 {
					return newPathError(
						ErrPathIndexOutOfBound,
						fmt.Errorf(
							"Index out of bound: %d",
							ind,
						),
					)
				}

				// If path needs to be reserved in the middle of the array, we should fill the gap with Null
				for j := v.Len(); j < ind; j++ {
					v.Append(Null)
				}

				v.Append(must.NotFail(NewDocument()))

			default:
				return newPathError(
					ErrPathCannotCreateField,
					fmt.Errorf(
						"Cannot create field '%s' in element {%s: %s}",
						pathElem,
						path.Prefix(),
						FormatAnyValue(must.NotFail(doc.Get(path.Prefix()))),
					),
				)
			}

			next = must.NotFail(doc.GetByPath(insertedPath)).(*Document)

			continue
		}

		next = v
	}

	return nil
}
