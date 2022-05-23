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
	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../../bin/stringer -linecomment -type ErrorCode

// ErrorCode represents wire protocol error code.
type ErrorCode int32

const (
	errUnset = ErrorCode(0) // Unset

	// For ProtocolError only.
	errInternalError = ErrorCode(1) // InternalError

	ErrBadValue          = ErrorCode(2)     // BadValue
	ErrFailedToParse     = ErrorCode(9)     // FailedToParse
	ErrTypeMismatch      = ErrorCode(14)    // TypeMismatch
	ErrNamespaceNotFound = ErrorCode(26)    // NamespaceNotFound
	ErrNamespaceExists   = ErrorCode(48)    // NamespaceExists
	ErrCommandNotFound   = ErrorCode(59)    // CommandNotFound
	ErrInvalidNamespace  = ErrorCode(73)    // InvalidNamespace
	ErrNotImplemented    = ErrorCode(238)   // NotImplemented
	ErrSortBadValue      = ErrorCode(15974) // Location15974
	ErrInvalidArg        = ErrorCode(28667) // Location28667
	ErrSliceFirstArg     = ErrorCode(28724) // Location28724
	ErrProjectionInEx    = ErrorCode(31253) // Location31253
	ErrProjectionExIn    = ErrorCode(31254) // Location31254
	ErrRegexOptions      = ErrorCode(51075) // Location51075
	ErrRegexMissingParen = ErrorCode(51091) // Location51091
)

// ProtoErr represents protocol error type.
type ProtoErr interface {
	error
	// Code returns ErrorCode.
	Code() ErrorCode
	// Document returns *types.Document.
	Document() *types.Document
}

// ProtocolError converts any error to wire protocol error.
//
// Nil panics, *Error or *WriteError (possibly wrapped) is returned unwrapped with true,
// any other value is wrapped with InternalError and returned with false.
func ProtocolError(err error) (ProtoErr, bool) {
	if err == nil {
		panic("err is nil")
	}

	var e *Error
	if errors.As(err, &e) {
		return e, true
	}

	var writeErr *WriteErrors
	if errors.As(err, &writeErr) {
		return writeErr, true
	}

	return NewError(errInternalError, err).(*Error), false
}

// Error represents wire command protocol error.
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

// Code implements ProtoErr interface.
func (e *Error) Code() ErrorCode {
	return e.code
}

// Unwrap implements standard error unwrapping interface.
func (e *Error) Unwrap() error {
	return e.err
}

// Document returns wire protocol error document.
func (e *Error) Document() *types.Document {
	d := must.NotFail(types.NewDocument(
		"ok", float64(0),
		"errmsg", e.err.Error(),
	))
	if e.code != errUnset {
		must.NoError(d.Set("code", int32(e.code)))
		must.NoError(d.Set("codeName", e.code.String()))
	}
	return d
}

// WriteErrors represents slice of protocol write errors.
type WriteErrors []WriteError

// NewWriteErrorMsg creates new protocol write error with given ErrorCode and message.
func NewWriteErrorMsg(code ErrorCode, msg string) error {
	return &WriteErrors{{
		code: code,
		err:  errors.New(msg),
	}}
}

// Error implements error interface.
func (we *WriteErrors) Error() string {
	var err string
	for _, e := range *we {
		err += e.Error() + ","
	}

	return err
}

// Code implements ProtoErr interface.
func (we *WriteErrors) Code() ErrorCode {
	for _, e := range *we {
		return e.code
	}
	return errUnset
}

// Unwrap implements standard error unwrapping interface.
func (we *WriteErrors) Unwrap() error {
	for _, e := range *we {
		return &e
	}
	return nil
}

// Document implements ProtoErr interface..
func (we *WriteErrors) Document() *types.Document {
	errs := must.NotFail(types.NewArray())
	for _, e := range *we {
		must.NoError(errs.Append(e.Document()))
	}

	d := must.NotFail(types.NewDocument(
		"ok", float64(1),
		"writeErrors", errs,
	))
	return d
}

// WriteError represents protocol write error.
// It required to build the correct write error result.
type WriteError struct {
	code ErrorCode
	err  error
}

// Error implements error interface.
func (we *WriteError) Error() string {
	return fmt.Sprintf("%[1]s (%[1]d): %[2]v", we.code, we.err)
}

// Code implements ProtoErr interface.
func (we *WriteError) Code() ErrorCode {
	return we.code
}

// Unwrap implements standard error unwrapping interface.
func (we *WriteError) Unwrap() error {
	return we.err
}

// Document implements ProtoErr interface.
// Fields "code" and "errmsg" must always be filled in so that clients can parse the error message.
func (we *WriteError) Document() *types.Document {
	d := must.NotFail(types.NewDocument(
		"code", int32(we.code),
		"errmsg", we.err.Error(),
	))
	return d
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
	_ error = (*WriteError)(nil)
	_ error = (*WriteErrors)(nil)
)
