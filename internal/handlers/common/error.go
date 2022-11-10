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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
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

	// ErrInvalidID indicates that _id field is invalid.
	ErrInvalidID = ErrorCode(53) // InvalidID

	// ErrCommandNotFound indicates unknown command input.
	ErrCommandNotFound = ErrorCode(59) // CommandNotFound

	// ErrInvalidNamespace indicates that the collection name is invalid.
	ErrInvalidNamespace = ErrorCode(73) // InvalidNamespace

	// ErrOperationFailed indicates that the operation failed.
	ErrOperationFailed = ErrorCode(96) // OperationFailed

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

	// ErrBadRegexOption indicates bad regex option value passed.
	ErrBadRegexOption = ErrorCode(51108) // Location51108
)

// ErrInfo represents additional optional error information.
type ErrInfo struct {
	Argument string // command's argument, operator, or aggregation pipeline stage
}

// ProtoErr represents protocol error type.
type ProtoErr interface {
	error
	// Unwrap returns unwrapped error.
	Unwrap() error
	// Code returns ErrorCode.
	Code() ErrorCode
	// Document returns *types.Document.
	Document() *types.Document
	// Info returns *ErrInfo.
	Info() *ErrInfo
}

// ProtocolError converts any error to wire protocol error.
//
// Nil panics, *CommandError or *WriteErrors (possibly wrapped) is returned unwrapped with true,
// any other value is wrapped with InternalError and returned with false.
func ProtocolError(err error) (ProtoErr, bool) {
	if err == nil {
		panic("err is nil")
	}

	var e *CommandError
	if errors.As(err, &e) {
		return e, true
	}

	var writeErr *WriteErrors
	if errors.As(err, &writeErr) {
		return writeErr, true
	}

	e = NewCommandError(errInternalError, err).(*CommandError) //nolint:errorlint // false positive

	return e, false
}

// formatBitwiseOperatorErr formats protocol error for given internal error and bitwise operator.
// Mask value used in error message.
func formatBitwiseOperatorErr(err error, operator string, maskValue any) error {
	switch err {
	case errNotWholeNumber:
		return NewCommandErrorMsgWithArgument(
			ErrFailedToParse,
			fmt.Sprintf("Expected an integer: %s: %#v", operator, maskValue),
			operator,
		)

	case errNegativeNumber:
		if _, ok := maskValue.(float64); ok {
			return NewCommandErrorMsgWithArgument(
				ErrFailedToParse,
				fmt.Sprintf(`Expected a non-negative number in: %s: %.1f`, operator, maskValue),
				operator,
			)
		}

		return NewCommandErrorMsgWithArgument(
			ErrFailedToParse,
			fmt.Sprintf(`Expected a non-negative number in: %s: %v`, operator, maskValue),
			operator,
		)

	case errNotBinaryMask:
		return NewCommandErrorMsgWithArgument(
			ErrBadValue,
			fmt.Sprintf(`value takes an Array, a number, or a BinData but received: %s: %#v`, operator, maskValue),
			operator,
		)

	default:
		return err
	}
}

// CheckError checks error type and returns properly translated error.
func CheckError(err error) error {
	var ve *types.ValidationError

	if !errors.As(err, &ve) {
		return lazyerrors.Error(err)
	}

	switch ve.Code() {
	case types.ErrValidation, types.ErrIDNotFound:
		return NewCommandErrorMsg(ErrBadValue, err.Error())
	case types.ErrWrongIDType:
		return NewWriteErrorMsg(ErrInvalidID, err.Error())
	default:
		panic(fmt.Sprintf("Unknown error code: %v", ve.Code()))
	}
}
