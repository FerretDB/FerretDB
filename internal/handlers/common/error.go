// Copyright 2021 Baltoro OÜ.
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
	"fmt"

	"github.com/MangoDB-io/MangoDB/internal/types"
)

//go:generate ../../../bin/stringer -linecomment -type ErrorCode

type ErrorCode int32

const (
	ErrInternalError     = ErrorCode(1)     // InternalError
	ErrBadValue          = ErrorCode(2)     // BadValue
	ErrNamespaceNotFound = ErrorCode(26)    // NamespaceNotFound
	ErrCommandNotFound   = ErrorCode(59)    // CommandNotFound
	ErrNotImplemented    = ErrorCode(238)   // NotImplemented
	ErrRegexOptions      = ErrorCode(51075) // Location51075
)

type Error struct {
	Code ErrorCode
	Err  error
}

func NewError(code ErrorCode, err error) error {
	return Error{
		Code: code,
		Err:  err,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("%[1]s (%[1]d): %[2]v", e.Code, e.Err)
}

func (e Error) Document() types.Document {
	return types.MustMakeDocument(
		"ok", float64(0),
		"errmsg", e.Err.Error(),
		"code", int32(e.Code),
		"codeName", e.Code.String(),
	)
}
