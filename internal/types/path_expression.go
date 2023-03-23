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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"strings"
)

//go:generate ../../bin/stringer -linecomment -type FieldPathErrorCode

// FieldPathErrorCode represents FieldPath error code.
type FieldPathErrorCode int

const (
	_ FieldPathErrorCode = iota

	// ErrNotFieldPath indicates that given field is not path.
	ErrNotFieldPath

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
	return e.Error()
}

// Code returns the FieldPathError code.
func (e *FieldPathError) Code() FieldPathErrorCode {
	return e.code
}

// GetFieldPath path gets the path using dollar literal field path.
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
