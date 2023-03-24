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

// Package commonerrors providers errors shared by all handlers.
package commonerrors

import (
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
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

	// ErrCursorNotFound indicates that cursor is not found.
	ErrCursorNotFound = ErrorCode(43) // CursorNotFound

	// ErrNamespaceExists indicates that the collection already exists.
	ErrNamespaceExists = ErrorCode(48) // NamespaceExists

	// ErrInvalidID indicates that _id field is invalid.
	ErrInvalidID = ErrorCode(53) // InvalidID

	// ErrEmptyName indicates that the field name is empty.
	ErrEmptyName = ErrorCode(56) // EmptyName

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

	// ErrMechanismUnavailable indicates unsupported authentication mechanism.
	ErrMechanismUnavailable = ErrorCode(334) // MechanismUnavailable

	// ErrDuplicateKey indicates duplicate key violation.
	ErrDuplicateKey = ErrorCode(11000) // Location11000

	// ErrStageGroupInvalidFields indicates group's fields must be an object.
	ErrStageGroupInvalidFields = ErrorCode(15947) // Location15947

	// ErrStageGroupID indicates _id for a group can only be specified once.
	ErrStageGroupID = ErrorCode(15948) // Location15948

	// ErrStageGroupMissingID indicates that group is missing an _id.
	ErrStageGroupMissingID = ErrorCode(15955) // Location15955

	// ErrStageLimitNotPositive indicated that limit field is not positive.
	ErrStageLimitNotPositive = ErrorCode(15958) // Location15958

	// ErrMatchBadExpression indicates match filter is not object.
	ErrMatchBadExpression = ErrorCode(15959) // Location15959

	// ErrSortBadExpression indicates sort expression is not object.
	ErrSortBadExpression = ErrorCode(15973) // Location15973

	// ErrSortBadValue indicates bad value in sort input.
	ErrSortBadValue = ErrorCode(15974) // Location15974

	// ErrSortBadOrder indicates bad sort order input.
	ErrSortBadOrder = ErrorCode(15975) // Location15975

	// ErrSortMissingKey indicates sort stage is missing sort key.
	ErrSortMissingKey = ErrorCode(15976) // Location15976

	// ErrPathContainsEmptyElement indicates that the path contains an empty element.
	ErrPathContainsEmptyElement = ErrorCode(15998) // Location15998

	// ErrGroupInvalidFieldPath indicates invalid path is given for group _id.
	ErrGroupInvalidFieldPath = ErrorCode(16872) // Location16872

	// ErrInvalidArg indicates invalid argument in projection document.
	ErrInvalidArg = ErrorCode(28667) // Location28667

	// ErrSliceFirstArg for $slice indicates that the first argument is not an array.
	ErrSliceFirstArg = ErrorCode(28724) // Location28724

	// ErrProjectionInEx for $elemMatch indicates that inclusion statement found
	// while projection document already marked as exclusion.
	ErrProjectionInEx = ErrorCode(31253) // Location31253

	// ErrProjectionExIn for $elemMatch indicates that exclusion statement found
	// while projection document already marked as inclusion.
	ErrProjectionExIn = ErrorCode(31254) // Location31254

	// ErrStageCountNonString indicates that $count aggregation stage expected string.
	ErrStageCountNonString = ErrorCode(40156) // Location40156

	// ErrStageCountNonEmptyString indicates that $count aggregation stage expected non-empty string.
	ErrStageCountNonEmptyString = ErrorCode(40157) // Location40157

	// ErrStageCountBadPrefix indicates that $count stage cannot have invalid prefix.
	ErrStageCountBadPrefix = ErrorCode(40158) // Location40158

	// ErrStageCountBadValue indicates that $count stage contains invalid value.
	ErrStageCountBadValue = ErrorCode(40160) // Location40160

	// ErrStageGroupMultipleAccumulator indicates that group field must specify one accumulator.
	ErrStageGroupMultipleAccumulator = ErrorCode(40238) // Location40238

	// ErrStageInvalid indicates invalid aggregation pipeline stage.
	ErrStageInvalid = ErrorCode(40323) // Location40323

	// ErrStageGroupInvalidAccumulator indicates invalid accumulator field.
	ErrStageGroupInvalidAccumulator = ErrorCode(40234) // Location40234

	// ErrEmptyFieldPath indicates that the field path is empty.
	ErrEmptyFieldPath = ErrorCode(40352) // Location40352

	// ErrMissingField indicates that the required field in document is missing.
	ErrMissingField = ErrorCode(40414) // Location40414

	// ErrFailedToParseInput indicates invalid input (absent or malformed fields).
	ErrFailedToParseInput = ErrorCode(40415) // Location40415

	// ErrFreeMonitoringDisabled indicates that free monitoring is disabled
	// by command-line or config file.
	ErrFreeMonitoringDisabled = ErrorCode(50840) // Location50840

	// ErrValueNegative indicates that value must not be negative.
	ErrValueNegative = ErrorCode(51024) // Location51024

	// ErrRegexOptions indicates regex options error.
	ErrRegexOptions = ErrorCode(51075) // Location51075

	// ErrRegexMissingParen indicates missing parentheses in regex expression.
	ErrRegexMissingParen = ErrorCode(51091) // Location51091

	// ErrBadRegexOption indicates bad regex option value passed.
	ErrBadRegexOption = ErrorCode(51108) // Location51108

	// ErrDuplicateField indicates duplicate field is specified.
	ErrDuplicateField = ErrorCode(4822819) // Location4822819

	// ErrStageLimitInvalidArg indicates invalid argument for the aggregation $limit stage.
	ErrStageLimitInvalidArg = ErrorCode(5107201) // Location5107201
)

// ErrInfo represents additional optional error information.
type ErrInfo struct {
	Argument string // command's argument, operator, or aggregation pipeline stage that caused an error
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

	var validationErr *wire.ValidationError
	if errors.As(err, &validationErr) {
		return NewCommandError(ErrBadValue, err).(*CommandError), true //nolint:errorlint // false positive
	}

	e = NewCommandError(errInternalError, err).(*CommandError) //nolint:errorlint // false positive

	return e, false
}

// CheckError checks error type and returns properly translated error.
func CheckError(err error) error {
	var ve *types.ValidationError

	if !errors.As(err, &ve) {
		return lazyerrors.Error(err)
	}

	switch ve.Code() {
	case types.ErrValidation, types.ErrIDNotFound:
		return NewCommandErrorMsg(ErrBadValue, ve.Error())
	case types.ErrWrongIDType:
		return NewWriteErrorMsg(ErrInvalidID, ve.Error())
	default:
		panic(fmt.Sprintf("Unknown error code: %v", ve.Code()))
	}
}
