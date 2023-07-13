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

package backends

import (
	"errors"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
)

//go:generate ../../bin/stringer -type ErrorCode

// ErrorCode represent a backend error code.
type ErrorCode int

// Error codes.
const (
	_ ErrorCode = iota

	ErrorCodeDatabaseDoesNotExist

	ErrorCodeCollectionDoesNotExist
	ErrorCodeCollectionAlreadyExists
	ErrorCodeCollectionNameIsInvalid
)

// Error represents a backend error returned by all Backend, Database and Collection methods.
type Error struct {
	// this internal error can't be accessed by the caller;
	// it may be nil
	err error

	code ErrorCode
}

// NewError creates a new backend error.
//
// Code must not be 0. Err may be nil.
func NewError(code ErrorCode, err error) *Error {
	if code == 0 {
		panic("backends.NewError: code must not be 0")
	}

	return &Error{
		code: code,
		err:  err,
	}
}

// Code returns the error code.
func (err *Error) Code() ErrorCode {
	return err.code
}

// There is intentionally no method to return the internal error.

// Error implements error interface.
func (err *Error) Error() string {
	return fmt.Sprintf("%s: %v", err.code, err.err)
}

// ErrorCodeIs returns true if err is *Error with one of the given error codes.
func ErrorCodeIs(err error, code ErrorCode, codes ...ErrorCode) bool {
	e, ok := err.(*Error) //nolint:errorlint // do not inspect error chain
	if !ok {
		return false
	}

	return e.code == code || slices.Contains(codes, e.code)
}

// checkError enforces backend interfaces contracts.
//
// Err must be nil, *Error, or some other opaque error.
// *Error values can't be wrapped or be present anywhere in the error chain.
// If err is *Error, it must have one of the given error codes.
// If that's not the case, checkError panics in debug builds.
//
// It does nothing in non-debug builds.
func checkError(err error, codes ...ErrorCode) {
	if !debugbuild.Enabled {
		return
	}

	if err == nil {
		return
	}

	e, ok := err.(*Error) //nolint:errorlint // do not inspect error chain
	if !ok {
		if errors.As(err, &e) {
			panic(fmt.Sprintf("error should not be wrapped: %v", err))
		}

		return
	}

	if e.code == 0 {
		panic(fmt.Sprintf("error code is 0: %v", err))
	}

	if len(codes) == 0 {
		panic(fmt.Sprintf("no allowed error codes: %v", err))
	}

	if !slices.Contains(codes, e.code) {
		panic(fmt.Sprintf("error code is not in %v: %v", codes, err))
	}
}

// check interfaces
var (
	_ error = (*Error)(nil)
)
