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
	"github.com/FerretDB/FerretDB/internal/util/must"
)

//go:generate ../../../bin/stringer -linecomment -type ErrorCode

// ErrorCode represents wire protocol error code.
type ErrorCode int32

const (
	errUnset = ErrorCode(0) // Unset

	// For ProtocolError only.
	errInternalError = ErrorCode(1) // InternalError

	// ErrBadValue indicates wrong input.
	ErrBadValue = ErrorCode(2) // BadValue

	// ErrFailedToParse indicates user input parsing failure.
	ErrFailedToParse = ErrorCode(9) // FailedToParse

	// ErrTypeMismatch for $sort indicates that the expression in the $sort is not an object.
	ErrTypeMismatch = ErrorCode(14) // TypeMismatch

	// ErrNamespaceNotFound indicates that a collection is not found.
	ErrNamespaceNotFound = ErrorCode(26) // NamespaceNotFound

	// ErrUnsuitableValueType indicates that field could not be created for given value.
	ErrUnsuitableValueType = ErrorCode(28) // UnsuitableValueType

	// ErrConflictingUpdateOperators indicates that $set, $inc or $setOnInsert were used together.
	ErrConflictingUpdateOperators = ErrorCode(40) // ConflictingUpdateOperators

	// ErrNamespaceExists indicates that the collection already exists.
	ErrNamespaceExists = ErrorCode(48) // NamespaceExists

	// ErrCommandNotFound indicates unknown command input.
	ErrCommandNotFound = ErrorCode(59) // CommandNotFound

	// ErrInvalidNamespace indicates that the collection name is invalid.
	ErrInvalidNamespace = ErrorCode(73) // InvalidNamespace

	// ErrDocumentValidationFailure indicates that document validation failed.
	ErrDocumentValidationFailure = ErrorCode(121) // DocumentValidationFailure

	// ErrNotImplemented indicates that a flag or command is not implemented.
	ErrNotImplemented = ErrorCode(238) // NotImplemented

	// ErrFailedToParseInput indicates invalid input (absent or malformed fields).
	ErrFailedToParseInput = ErrorCode(40415) // Location40415

	// ErrSortBadValue indicates bad value in sort input.
	ErrSortBadValue = ErrorCode(15974) // Location15974

	// ErrSortBadOrder indicates bad sort order input.
	ErrSortBadOrder = ErrorCode(15975) // Location15975

	// ErrInvalidArg indicates invalid argument in projection document.
	ErrInvalidArg = ErrorCode(28667) // Location28667

	// ErrSliceFirstArg for $slice indicates that the first argument is not an array.
	ErrSliceFirstArg = ErrorCode(28724) // Location28724

	// ErrProjectionInEx for $elemMatch indicates that inclusion statement found
	// while projection document already marked as exlusion.
	ErrProjectionInEx = ErrorCode(31253) // Location31253

	// ErrProjectionExIn for $elemMatch indicates that exlusion statement found
	// while projection document already marked as inclusion.
	ErrProjectionExIn = ErrorCode(31254) // Location31254

	// ErrMissingField indicates that the required field in document is missing.
	ErrMissingField = ErrorCode(40414) // Location40414

	// ErrFreeMonitoringDisabled indicates that free monitoring is disabled
	// by command-line or config file.
	ErrFreeMonitoringDisabled = ErrorCode(50840) // Location50840

	// ErrRegexOptions indicates regex options error.
	ErrRegexOptions = ErrorCode(51075) // Location51075

	// ErrRegexMissingParen indicates missing parentheses in regex expression.
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
// Nil panics, *Error or *WriteErrors (possibly wrapped) is returned unwrapped with true,
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

// CommandError represents wire protocol command error.
type CommandError = Error

// Error is a deprecated name for CommandError; instead, use the later version in the new code.
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

// WriteErrors represents a slice of protocol write errors.
// It could be returned for Update, Insert, Delete, and Replace operations.
type WriteErrors []writeError

// NewWriteErrorMsg creates a new protocol write error with given ErrorCode and message.
func NewWriteErrorMsg(code ErrorCode, msg string) error {
	return &WriteErrors{{
		code: code,
		err:  msg,
	}}
}

// Error implements error interface.
func (we *WriteErrors) Error() string {
	var err string
	for _, e := range *we {
		err += e.err + ","
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

// Unwrap implements a standard error unwrapping interface.
func (we *WriteErrors) Unwrap() error {
	for _, e := range *we {
		return errors.New(e.err)
	}
	return nil
}

// Document implements ProtoErr interface.
func (we *WriteErrors) Document() *types.Document {
	errs := must.NotFail(types.NewArray())
	for _, e := range *we {
		doc := must.NotFail(types.NewDocument())

		if e.index != nil {
			must.NoError(doc.Set("index", *e.index))
		}

		// Fields "code" and "errmsg" must always be filled in so that clients can parse the error message.
		// Otherwise, the mongo client would parse it as a CommandError.
		must.NoError(doc.Set("code", int32(e.code)))
		must.NoError(doc.Set("errmsg", e.err))

		must.NoError(errs.Append(doc))
	}

	// "writeErrors" field must be present in the result document so that clients can parse it as WriteErrors.
	d := must.NotFail(types.NewDocument(
		"ok", float64(1),
		"writeErrors", errs,
	))
	return d
}

// Append converts the err to the writeError type and
// appends it to WriteErrors. The index value is an
// index of the query with error.
func (we *WriteErrors) Append(err error, index int32) {
	var writeErr *writeError
	var cmdErr *CommandError

	switch {
	case errors.As(err, &writeErr):
		writeErr.index = &index
		*we = append(*we, *writeErr)

		return

	case errors.As(err, &cmdErr):
		*we = append(*we, writeError{err: cmdErr.Unwrap().Error(), code: cmdErr.code, index: &index})

		return
	}

	*we = append(*we, writeError{err: err.Error(), code: errInternalError, index: &index})
}

// writeError represents protocol write error.
// It required to build the correct write error result.
// The index field is optional and won't be used if it's nil.
type writeError struct {
	code  ErrorCode
	err   string
	index *int32
}

// Error returns the string that contains
// an error message.
func (we *writeError) Error() string {
	return we.err
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
	_ ProtoErr = (*Error)(nil)
	_ ProtoErr = (*WriteErrors)(nil)
)
