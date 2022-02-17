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

package common

import (
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/types"
)

//go:generate ../../../bin/stringer -linecomment -type ErrorCode

// ErrorCode represents wire protocol error code.
type ErrorCode int32

const (
	// For ProtocolError only.
	errInternalError = ErrorCode(1) // InternalError

	ErrBadValue          = ErrorCode(2)     // BadValue
	ErrNamespaceNotFound = ErrorCode(26)    // NamespaceNotFound
	ErrNamespaceExists   = ErrorCode(48)    // NamespaceExists
	ErrCommandNotFound   = ErrorCode(59)    // CommandNotFound
	ErrNotImplemented    = ErrorCode(238)   // NotImplemented
	ErrRegexOptions      = ErrorCode(51075) // Location51075
)

// Error represents wire protocol error.
type Error struct {
	code ErrorCode
	err  error
}

// NewError creates a new wire protocol error.
//
// Code can't be zero, err can't be nil.
func NewError(code ErrorCode, err error) error {
	if code == 0 {
		panic("code is 0")
	}
	if err == nil {
		panic("err is nil")
	}
	return &Error{
		code: code,
		err:  err,
	}
}

// NewErrorMsg is variant for NewError with error string.
//
// Code can't be zero, err can't be empty.
func NewErrorMsg(code ErrorCode, msg string) error {
	return NewError(code, errors.New(msg))
}

// Error implements error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%[1]s (%[1]d): %[2]v", e.code, e.err)
}

// Unwrap implements standard error unwrapping interface.
func (e *Error) Unwrap() error {
	return e.err
}

// Document returns wire protocol error document.
func (e *Error) Document() *types.Document {
	return types.MustNewDocument(
		"ok", float64(0),
		"errmsg", e.err.Error(),
		"code", int32(e.code),
		"codeName", e.code.String(),
	)
}

// ProtocolError converts any error to wire protocol error.
//
// Nil panics, *Error (possibly wrapped) is returned unwrapped with true,
// any other value is wrapped with InternalError and returned with false.
func ProtocolError(err error) (*Error, bool) {
	if err == nil {
		panic("err is nil")
	}

	var e *Error
	if errors.As(err, &e) {
		return e, true
	}

	return NewError(errInternalError, err).(*Error), false
}

// check interfaces
var (
	_ error = (*Error)(nil)
)
