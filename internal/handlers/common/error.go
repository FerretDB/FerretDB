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

// Package common provides common code for all handlers.
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
	ErrFailedToParse     = ErrorCode(9)     // FailedToParse
	ErrTypeMismatch      = ErrorCode(14)    // TypeMismatch
	ErrNamespaceNotFound = ErrorCode(26)    // NamespaceNotFound
	ErrNamespaceExists   = ErrorCode(48)    // NamespaceExists
	ErrCommandNotFound   = ErrorCode(59)    // CommandNotFound
	ErrNotImplemented    = ErrorCode(238)   // NotImplemented
	ErrSortBadValue      = ErrorCode(15974) // Location15974
	ErrProjectionInEx    = ErrorCode(31253) // Location31253
	ErrProjectionExIn    = ErrorCode(31254) // Location31254
	ErrRegexOptions      = ErrorCode(51075) // Location51075
)

// Error represents wire protocol error.
type Error struct {
	err  error
	code ErrorCode
}

// There should not be NewError function variant that accepts printf-like format specifiers.
// Let the caller do safe formatting.

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

// Code returns error code.
func (e *Error) Code() ErrorCode {
	return e.code
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

// formatBitwiseOperatorErr formats protocol error for given internal error and bitwise operator.
// Mask value used in error message.
func formatBitwiseOperatorErr(err error, operator string, maskValue any) error {
	switch err {
	case errNotWholeNumber:
		return NewErrorMsg(
			ErrFailedToParse,
			fmt.Sprintf("Expected an integer: %s: %#v", operator, maskValue),
		)

	case errNegativeNumber:
		if _, ok := maskValue.(float64); ok {
			return NewErrorMsg(
				ErrFailedToParse,
				fmt.Sprintf(`Expected a positive number in: %s: %.1f`, operator, maskValue),
			)
		}
		return NewErrorMsg(
			ErrFailedToParse,
			fmt.Sprintf(`Expected a positive number in: %s: %v`, operator, maskValue),
		)

	case errNotBinaryMask:
		return NewErrorMsg(
			ErrBadValue,
			fmt.Sprintf(`value takes an Array, a number, or a BinData but received: %s: %#v`, operator, maskValue),
		)

	default:
		return err
	}
}

// check interfaces
var (
	_ error = (*Error)(nil)
)
