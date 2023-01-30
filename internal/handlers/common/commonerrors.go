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

import "github.com/FerretDB/FerretDB/internal/handlers/commonerrors"

type (
	// ErrorCode should not be used in the new or updated code.
	// Deprecated: use commonerrors types instead.
	ErrorCode = commonerrors.ErrorCode

	// WriteErrors should not be used in the new or updated code.
	// Deprecated: use commonerrors types instead.
	WriteErrors = commonerrors.WriteErrors
)

// Deprecated: use commonerrors functions instead; do not add new functions there.
var (
	CheckError                     = commonerrors.CheckError
	NewCommandError                = commonerrors.NewCommandError
	NewCommandErrorMsg             = commonerrors.NewCommandErrorMsg
	NewCommandErrorMsgWithArgument = commonerrors.NewCommandErrorMsgWithArgument
	NewWriteErrorMsg               = commonerrors.NewWriteErrorMsg
	ProtocolError                  = commonerrors.ProtocolError
)

// Deprecated: use commonerrors constant instead; do not add new errors there.
const (
	ErrBadValue                   = commonerrors.ErrBadValue
	ErrFailedToParse              = commonerrors.ErrFailedToParse
	ErrTypeMismatch               = commonerrors.ErrTypeMismatch
	ErrNamespaceNotFound          = commonerrors.ErrNamespaceNotFound
	ErrUnsuitableValueType        = commonerrors.ErrUnsuitableValueType
	ErrConflictingUpdateOperators = commonerrors.ErrConflictingUpdateOperators
	ErrNamespaceExists            = commonerrors.ErrNamespaceExists
	ErrInvalidID                  = commonerrors.ErrInvalidID
	ErrEmptyName                  = commonerrors.ErrEmptyName
	ErrCommandNotFound            = commonerrors.ErrCommandNotFound
	ErrInvalidNamespace           = commonerrors.ErrInvalidNamespace
	ErrOperationFailed            = commonerrors.ErrOperationFailed
	ErrDocumentValidationFailure  = commonerrors.ErrDocumentValidationFailure
	ErrNotImplemented             = commonerrors.ErrNotImplemented
	ErrMechanismUnavailable       = commonerrors.ErrMechanismUnavailable
	ErrSortBadValue               = commonerrors.ErrSortBadValue
	ErrSortBadOrder               = commonerrors.ErrSortBadOrder
	ErrInvalidArg                 = commonerrors.ErrInvalidArg
	ErrSliceFirstArg              = commonerrors.ErrSliceFirstArg
	ErrProjectionInEx             = commonerrors.ErrProjectionInEx
	ErrProjectionExIn             = commonerrors.ErrProjectionExIn
	ErrEmptyFieldPath             = commonerrors.ErrEmptyFieldPath
	ErrMissingField               = commonerrors.ErrMissingField
	ErrFailedToParseInput         = commonerrors.ErrFailedToParseInput
	ErrFreeMonitoringDisabled     = commonerrors.ErrFreeMonitoringDisabled
	ErrRegexOptions               = commonerrors.ErrRegexOptions
	ErrRegexMissingParen          = commonerrors.ErrRegexMissingParen
	ErrBadRegexOption             = commonerrors.ErrBadRegexOption
)
