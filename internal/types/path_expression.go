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
	"strconv"
	"strings"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../bin/stringer -linecomment -type FieldPathErrorCode

// FieldPathErrorCode represents FieldPath error code.
type FieldPathErrorCode int

const (
	_ FieldPathErrorCode = iota

	// ErrNotFieldPath indicates that given field is not path.
	ErrNotFieldPath

	// ErrEmptyFieldPath indicates that field path is empty.
	ErrEmptyFieldPath

	// ErrInvalidFieldPath indicates that path is invalid.
	ErrInvalidFieldPath

	// ErrUndefinedVariable indicates that variable name is not defined.
	ErrUndefinedVariable

	// ErrEmptyVariable indicates that variable name is empty.
	ErrEmptyVariable
)

// FieldPathError describes an error that occurs getting path from field.
type FieldPathError struct {
	code FieldPathErrorCode
}

// newFieldPathError creates a new FieldPathError.
func newFieldPathError(code FieldPathErrorCode) error {
	return &FieldPathError{code: code}
}

// Error implements the error interface.
func (e *FieldPathError) Error() string {
	return e.code.String()
}

// Code returns the FieldPathError code.
func (e *FieldPathError) Code() FieldPathErrorCode {
	return e.code
}

// GetFieldPath gets the path using dollar literal field path.
func GetFieldPath(expression string) (Path, error) {
	var res Path

	var val string

	switch {
	case strings.HasPrefix(expression, "$$"):
		// two dollar signs indicates field has a variable.
		v := strings.TrimPrefix(expression, "$$")
		if v == "" {
			return res, newFieldPathError(ErrEmptyVariable)
		}

		if strings.HasPrefix(v, "$") {
			return res, newFieldPathError(ErrInvalidFieldPath)
		}

		// todo implement getting variable fetching
		return res, newFieldPathError(ErrUndefinedVariable)
	case strings.HasPrefix(expression, "$"):
		// one dollar signs indicates field is path.
		val = strings.TrimPrefix(expression, "$")

		if val == "" {
			return res, newFieldPathError(ErrEmptyFieldPath)
		}
	default:
		return res, newFieldPathError(ErrNotFieldPath)
	}

	var err error

	res, err = NewPathFromString(val)
	if err != nil {
		return res, lazyerrors.Error(err)
	}

	return res, nil
}

// GetFieldValue gets the value at the path using dollar literal field path.
func GetFieldValue(expression string, doc *Document) (any, error) {
	path, err := GetFieldPath(expression)
	if err != nil {
		return nil, err
	}

	if path.Len() == 1 {
		val, err := doc.Get(path.String())
		if err != nil {
			return nil, err
		}

		return val, nil
	}

	_, vals := getValuesAtSuffix(doc, path)

	if len(vals) == 0 {
		prefix := path.Prefix()
		if v, err := doc.Get(prefix); err == nil {
			if _, isArray := v.(*Array); isArray {
				// when the prefix is array, it returns empty array instead of null
				return MakeArray(0), nil
			}
		}

		return Null, nil
	}

	if len(vals) == 1 {
		return vals[0], nil
	}

	arr := MakeArray(len(vals))
	for _, v := range vals {
		arr.Append(v)
	}

	return NewArray(vals)
}

func getValuesAtSuffix(doc *Document, path Path) (string, []any) {
	// keys are each part of the path.
	keys := path.Slice()

	// vals are the field values found at each key of the path.
	vals := []any{doc}

	for _, key := range keys {
		// embeddedVals are the values found at current key.
		var embeddedVals []any

		for _, valAtKey := range vals {
			switch val := valAtKey.(type) {
			case *Document:
				embeddedVal, err := val.Get(key)
				if err != nil {
					// document does not contain key, so no embedded value was found.
					continue
				}

				// key exists in the document, add embedded value to next iteration.
				embeddedVals = append(embeddedVals, embeddedVal)
			case *Array:
				if index, err := strconv.Atoi(key); err == nil {
					// key is an integer, check if that integer is an index of the array.
					embeddedVal, err := val.Get(index)
					if err != nil {
						// index does not exist.
						continue
					}

					// key is the index of the array, add embedded value to the next iteration.
					embeddedVals = append(embeddedVals, embeddedVal)

					continue
				}

				// key was not an index, iterate array to get all documents that contain the key.
				for j := 0; j < val.Len(); j++ {
					valAtIndex := must.NotFail(val.Get(j))

					embeddedDoc, isDoc := valAtIndex.(*Document)
					if !isDoc {
						// the value is not a document, so it cannot contain the key.
						continue
					}

					embeddedVal, err := embeddedDoc.Get(key)
					if err != nil {
						// the document does not contain key, so no embedded value was found.
						continue
					}

					// key exists in the document, add embedded value to next iteration.
					embeddedVals = append(embeddedVals, embeddedVal)
				}

			default:
				// not a document or array, do nothing
			}
		}

		vals = embeddedVals
	}

	return path.Suffix(), vals
}
