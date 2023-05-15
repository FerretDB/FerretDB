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

//go:generate ../../bin/stringer -linecomment -type ErrorCode

type ErrorCode int

const (
	_ ErrorCode = iota

	ErrCollectionDoesNotExist  // collection does not exist
	ErrCollectionAlreadyExists // collection already exists
	ErrCollectionNameIsInvalid // collection name is invalid
)

type Error struct {
	code ErrorCode
	err  error
}

// NewError creates a new backend error wrapping another error.
func NewError(code ErrorCode, err error) *Error {
	if code == 0 {
		panic("backends.NewError: code must not be 0")
	}

	// TODO we might allow nil error if needed
	if err == nil {
		panic("backends.NewError: err must not be nil")
	}

	return &Error{
		code: code,
		err:  err,
	}
}

func (err *Error) Code() ErrorCode {
	return err.code
}

func (err *Error) Error() string {
	return fmt.Sprintf("backends.Error: %v", err.err)
}

func (err *Error) Unwrap() error {
	return err.err
}

func checkError(err error, codes ...ErrorCode) {
	if !debugbuild.Enabled {
		return
	}

	if err == nil {
		return
	}

	var e *Error
	if !errors.As(err, &e) {
		return
	}

	if e.code == 0 {
		panic(err)
	}

	if len(codes) == 0 {
		panic(err)
	}

	if !slices.Contains(codes, e.code) {
		panic(err)
	}
}

// check interfaces
var (
	_ error = (*Error)(nil)
)
