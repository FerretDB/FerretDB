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

	"github.com/FerretDB/FerretDB/internal/types"
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

	// ErrUnauthorized indicates that cursor is not authorized to access another namespace.
	ErrUnauthorized = ErrorCode(13) // Unauthorized

	// ErrTypeMismatch for $sort indicates that the expression in the $sort is not an object.
	ErrTypeMismatch = ErrorCode(14) // TypeMismatch

	// ErrAuthenticationFailed indicates failed authentication.
	ErrAuthenticationFailed = ErrorCode(18) // AuthenticationFailed

	// ErrIllegalOperation indicated that operation is illegal.
	ErrIllegalOperation = ErrorCode(20) // IllegalOperation

	// ErrNamespaceNotFound indicates that a collection is not found.
	ErrNamespaceNotFound = ErrorCode(26) // NamespaceNotFound

	// ErrIndexNotFound indicates that a index is not found for the given name.
	ErrIndexNotFound = ErrorCode(27) // IndexNotFound

	// ErrUnsuitableValueType indicates that field could not be created for given value.
	ErrUnsuitableValueType = ErrorCode(28) // PathNotViable

	// ErrConflictingUpdateOperators indicates that $set, $inc or $setOnInsert were used together.
	ErrConflictingUpdateOperators = ErrorCode(40) // ConflictingUpdateOperators

	// ErrCursorNotFound indicates that cursor is not found.
	ErrCursorNotFound = ErrorCode(43) // CursorNotFound

	// ErrNamespaceExists indicates that the collection already exists.
	ErrNamespaceExists = ErrorCode(48) // NamespaceExists

	// ErrDollarPrefixedFieldName indicates the field name is prefixed with $.
	ErrDollarPrefixedFieldName = ErrorCode(52) // DollarPrefixedFieldName

	// ErrInvalidID indicates that _id field is invalid.
	ErrInvalidID = ErrorCode(53) // InvalidID

	// ErrEmptyName indicates that the field name is empty.
	ErrEmptyName = ErrorCode(56) // EmptyFieldName

	// ErrCommandNotFound indicates unknown command input.
	ErrCommandNotFound = ErrorCode(59) // CommandNotFound

	// ErrImmutableField indicates that _id field is immutable.
	ErrImmutableField = ErrorCode(66) // ImmutableField

	// ErrCannotCreateIndex indicates that index creation process failed because some data are not valid.
	ErrCannotCreateIndex = ErrorCode(67) // CannotCreateIndex

	// ErrIndexAlreadyExists indicates that identical index already exists.
	ErrIndexAlreadyExists = ErrorCode(68) // IndexAlreadyExists

	// ErrInvalidOptions indicates that _id index cannot be deleted.
	ErrInvalidOptions = ErrorCode(72) // InvalidOptions

	// ErrInvalidNamespace indicates that the collection name is invalid.
	ErrInvalidNamespace = ErrorCode(73) // InvalidNamespace

	// ErrIndexOptionsConflict indicates that index build process failed due to options conflict.
	ErrIndexOptionsConflict = ErrorCode(85) // IndexOptionsConflict

	// ErrIndexKeySpecsConflict indicates that index build process failed due to key specs conflict.
	ErrIndexKeySpecsConflict = ErrorCode(86) // IndexKeySpecsConflict

	// ErrOperationFailed indicates that the operation failed.
	ErrOperationFailed = ErrorCode(96) // OperationFailed

	// ErrDocumentValidationFailure indicates that document validation failed.
	ErrDocumentValidationFailure = ErrorCode(121) // DocumentValidationFailure

	// ErrInvalidIndexSpecificationOption indicates that the index option is invalid.
	ErrInvalidIndexSpecificationOption = ErrorCode(197) // InvalidIndexSpecificationOption

	// ErrInvalidPipelineOperator indicates that provided aggregation operator is invalid.
	ErrInvalidPipelineOperator = ErrorCode(168) // InvalidPipelineOperator

	// ErrClientMetadataCannotBeMutated indicates that client metadata cannot be mutated.
	ErrClientMetadataCannotBeMutated = ErrorCode(186) // ClientMetadataCannotBeMutated

	// ErrNotImplemented indicates that a flag or command is not implemented.
	ErrNotImplemented = ErrorCode(238) // NotImplemented

	// ErrIndexesWrongType indicates that indexes parameter has wrong type.
	ErrIndexesWrongType = ErrorCode(10065) // Location10065

	// ErrDuplicateKeyInsert indicates duplicate key violation on inserting document.
	ErrDuplicateKeyInsert = ErrorCode(11000) // Location11000

	// ErrSetBadExpression indicates set expression is not object.
	ErrSetBadExpression = ErrorCode(40272) // Location40272

	// ErrStageGroupInvalidFields indicates group's fields must be an object.
	ErrStageGroupInvalidFields = ErrorCode(15947) // Location15947

	// ErrStageGroupID indicates _id for a group can only be specified once.
	ErrStageGroupID = ErrorCode(15948) // Location15948

	// ErrStageGroupMissingID indicates that group is missing an _id.
	ErrStageGroupMissingID = ErrorCode(15955) // Location15955

	// ErrStageLimitZero indicated that limit field is not positive.
	ErrStageLimitZero = ErrorCode(15958) // Location15958

	// ErrMatchBadExpression indicates match filter is not object.
	ErrMatchBadExpression = ErrorCode(15959) // Location15959

	// ErrProjectBadExpression indicates projection value wrong type.
	ErrProjectBadExpression = ErrorCode(15969) // Location15969

	// ErrSortBadExpression indicates sort expression is not object.
	ErrSortBadExpression = ErrorCode(15973) // Location15973

	// ErrSortBadValue indicates bad value in sort input.
	ErrSortBadValue = ErrorCode(15974) // Location15974

	// ErrSortBadOrder indicates bad sort order input.
	ErrSortBadOrder = ErrorCode(15975) // Location15975

	// ErrSortMissingKey indicates sort stage is missing sort key.
	ErrSortMissingKey = ErrorCode(15976) // Location15976

	// ErrGroupDuplicateFieldName indicates that duplicate field name is specified.
	ErrGroupDuplicateFieldName = ErrorCode(16406) // Location16406

	// ErrStageUnwindWrongType indicates $unwind stage argument has unexpected type.
	ErrStageUnwindWrongType = ErrorCode(15981) // Location15981

	// ErrExpressionWrongLenOfFields indicates that aggregation expression has too much fields.
	ErrExpressionWrongLenOfFields = ErrorCode(15983) // Location15983

	// ErrPathContainsEmptyElement indicates that the path contains an empty element.
	ErrPathContainsEmptyElement = ErrorCode(15998) // Location15998

	// ErrOperatorWrongLenOfArgs indicates that aggregation operator contains
	// wrong amount of arguments.
	ErrOperatorWrongLenOfArgs = ErrorCode(16020) // Location16020

	// ErrFieldPathInvalidName indicates that FieldPath is invalid.
	ErrFieldPathInvalidName = ErrorCode(16410) // Location16410

	// ErrGroupInvalidFieldPath indicates invalid path is given for group _id.
	ErrGroupInvalidFieldPath = ErrorCode(16872) // Location16872

	// ErrGroupUndefinedVariable indicates the variable is not defined.
	ErrGroupUndefinedVariable = ErrorCode(17276) // Location17276

	// ErrInvalidArg indicates invalid argument in projection document.
	ErrInvalidArg = ErrorCode(28667) // Location28667

	// ErrSliceFirstArg for $slice indicates that the first argument is not an array.
	ErrSliceFirstArg = ErrorCode(28724) // Location28724

	// ErrStageUnsetNoPath indicates that $unwind aggregation stage is empty.
	ErrStageUnsetNoPath = ErrorCode(31119) // Location31119

	// ErrStageUnsetArrElementInvalidType indicates that $unset stage arguments has array with unexpected type.
	ErrStageUnsetArrElementInvalidType = ErrorCode(31120) // Location31120

	// ErrStageUnsetInvalidType indicates that $unset stage arguments has unexpected type.
	ErrStageUnsetInvalidType = ErrorCode(31002) // Location31002

	// ErrStageUnwindNoPath indicates that $unwind aggregation stage is empty.
	ErrStageUnwindNoPath = ErrorCode(28812) // Location28812

	// ErrStageUnwindNoPrefix indicates that $unwind aggregation stage doesn't include '$' prefix.
	ErrStageUnwindNoPrefix = ErrorCode(28818) // Location28818

	// ErrUnsetPathCollision indicates that an $unset path creates collision at another path in arguments.
	ErrUnsetPathCollision = ErrorCode(31249) // Location31249

	// ErrUnsetPathOverwrite indicates that an $unset path have overwrites another path in arguments.
	ErrUnsetPathOverwrite = ErrorCode(31250) // Location31250

	// ErrProjectionInEx for $elemMatch indicates that inclusion statement found
	// while projection document already marked as exclusion.
	ErrProjectionInEx = ErrorCode(31253) // Location31253

	// ErrProjectionExIn for $elemMatch indicates that exclusion statement found
	// while projection document already marked as inclusion.
	ErrProjectionExIn = ErrorCode(31254) // Location31254

	// ErrAggregatePositionalProject indicates that positional projection cannot be used in aggregation.
	ErrAggregatePositionalProject = ErrorCode(31324) // Location31324

	// ErrAggregateInvalidExpression indicates that projection expression does not exist.
	ErrAggregateInvalidExpression = ErrorCode(31325) // Location31325

	// ErrWrongPositionalOperatorLocation indicates that there can only be one positional
	// operator at the end.
	ErrWrongPositionalOperatorLocation = ErrorCode(31394) // Location31394

	// ErrExclusionPositionalProjection indicates that exclusion cannot use positional projection.
	ErrExclusionPositionalProjection = ErrorCode(31395) // Location31395

	// ErrStageCountNonString indicates that $count aggregation stage expected string.
	ErrStageCountNonString = ErrorCode(40156) // Location40156

	// ErrStageCountNonEmptyString indicates that $count aggregation stage expected non-empty string.
	ErrStageCountNonEmptyString = ErrorCode(40157) // Location40157

	// ErrStageCountBadPrefix indicates that $count stage cannot have invalid prefix.
	ErrStageCountBadPrefix = ErrorCode(40158) // Location40158

	// ErrStageCountBadValue indicates that $count stage contains invalid value.
	ErrStageCountBadValue = ErrorCode(40160) // Location40160

	// ErrAddFieldsExpressionWrongAmountOfArgs indicates that $addFields stage expression contain invalid
	// amount of arguments.
	ErrAddFieldsExpressionWrongAmountOfArgs = ErrorCode(40181) // Location40181

	// ErrStageGroupUnaryOperator indicates that $sum is a unary operator.
	ErrStageGroupUnaryOperator = ErrorCode(40237) // Location40237

	// ErrStageGroupMultipleAccumulator indicates that group field must specify one accumulator.
	ErrStageGroupMultipleAccumulator = ErrorCode(40238) // Location40238

	// ErrStageGroupInvalidAccumulator indicates invalid accumulator field.
	ErrStageGroupInvalidAccumulator = ErrorCode(40234) // Location40234

	// ErrStageInvalid indicates invalid aggregation pipeline stage.
	ErrStageInvalid = ErrorCode(40323) // Location40323

	// ErrEmptyFieldPath indicates that the field path is empty.
	ErrEmptyFieldPath = ErrorCode(40352) // Location40352

	// ErrInvalidFieldPath indicates that the field path is not valid.
	ErrInvalidFieldPath = ErrorCode(40353) // Location40353

	// ErrMissingField indicates that the required field in document is missing.
	ErrMissingField = ErrorCode(40414) // Location40414

	// ErrFailedToParseInput indicates invalid input (absent or malformed fields).
	ErrFailedToParseInput = ErrorCode(40415) // Location40415

	// ErrCollStatsIsNotFirstStage indicates that $collStats must be the first stage in the pipeline.
	ErrCollStatsIsNotFirstStage = ErrorCode(40602) // Location40602

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

	// ErrBadPositionalProjection indicates that positional operator could not find a matching element in the array.
	ErrBadPositionalProjection = ErrorCode(51246) // Location51246

	// ErrElementMismatchPositionalProjection indicates that unexpected element was present at projection path.
	ErrElementMismatchPositionalProjection = ErrorCode(51247) // Location51247

	// ErrEmptySubProject indicates that subprojection mustn't be empty.
	ErrEmptySubProject = ErrorCode(51270) // Location51270

	// ErrEmptyProject indicates that projection specification must have at least one field.
	ErrEmptyProject = ErrorCode(51272) // Location51272

	// ErrDuplicateField indicates duplicate field is specified.
	ErrDuplicateField = ErrorCode(4822819) // Location4822819

	// ErrStageSkipBadValue indicates that $skip stage contains invalid value.
	ErrStageSkipBadValue = ErrorCode(5107200) // Location5107200

	// ErrStageLimitInvalidArg indicates invalid argument for the aggregation $limit stage.
	ErrStageLimitInvalidArg = ErrorCode(5107201) // Location5107201

	// ErrStageCollStatsInvalidArg indicates invalid argument for the aggregation $collStats stage.
	ErrStageCollStatsInvalidArg = ErrorCode(5447000) // Location5447000
)

// ErrInfo represents additional optional error information.
type ErrInfo struct {
	Argument string // command's argument, operator, or aggregation pipeline stage that caused an error
}

// ProtoErr represents protocol error type.
type ProtoErr interface {
	// Error returns error representation for logging and debugging.
	error

	// Document returns error representation for returning to the client.
	Document() *types.Document

	// Info returns additional error information, or nil.
	Info() *ErrInfo
}

// ProtocolError converts any error to wire protocol error.
//
// Nil panics (it never should be passed),
// *CommandError or *WriteErrors (possibly wrapped) are returned unwrapped,
// *wire.ValidationError (possibly wrapped) is returned as CommandError with BadValue code,
// any other values (including lazy errors) are returned as CommandError with InternalError code.
func ProtocolError(err error) ProtoErr {
	if err == nil {
		panic("err is nil")
	}

	var commandErr *CommandError
	if errors.As(err, &commandErr) {
		return commandErr
	}

	var writeErr *WriteErrors
	if errors.As(err, &writeErr) {
		return writeErr
	}

	var validationErr *wire.ValidationError
	if errors.As(err, &validationErr) {
		//nolint:errorlint // only *CommandError could be returned
		return NewCommandError(ErrBadValue, err).(*CommandError)
	}

	//nolint:errorlint // only *CommandError could be returned
	return NewCommandError(errInternalError, err).(*CommandError)
}
