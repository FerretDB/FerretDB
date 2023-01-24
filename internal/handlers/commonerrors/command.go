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

package commonerrors

import (
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// CommandError represents wire protocol command error.
type CommandError struct {
	err  error
	code ErrorCode
	info *ErrInfo
}

// There should not be NewCommandError function variant that accepts printf-like format specifiers.
// Let the caller do safe formatting.

// NewCommandError creates a new wire protocol error.
//
// Code shouldn't be zero, err can't be nil.
func NewCommandError(code ErrorCode, err error) error {
	if err == nil {
		panic("err is nil")
	}

	return &CommandError{
		code: code,
		err:  err,
	}
}

// NewCommandErrorMsg is variant for NewCommandError with error string.
//
// Code shouldn't be zero, err can't be empty.
func NewCommandErrorMsg(code ErrorCode, msg string) error {
	return NewCommandError(code, errors.New(msg))
}

// NewCommandErrorMsgWithArgument creates a new wire protocol error with an argument that caused the error.
func NewCommandErrorMsgWithArgument(code ErrorCode, msg string, argument string) error {
	return &CommandError{
		err:  errors.New(msg),
		code: code,
		info: &ErrInfo{
			Argument: argument,
		},
	}
}

// Error implements error interface.
func (e *CommandError) Error() string {
	return fmt.Sprintf("%[1]s (%[1]d): %[2]v", e.code, e.err)
}

// Unwrap implements ProtoErr interface.
func (e *CommandError) Unwrap() error {
	return e.err
}

// Code implements ProtoErr interface.
func (e *CommandError) Code() ErrorCode {
	return e.code
}

// Document implements ProtoErr interface.
func (e *CommandError) Document() *types.Document {
	d := must.NotFail(types.NewDocument(
		"ok", float64(0),
		"errmsg", e.err.Error(),
	))
	if e.code != errUnset {
		d.Set("code", int32(e.code))
		d.Set("codeName", e.code.String())
	}

	return d
}

// Info implements ProtoErr interface.
func (e *CommandError) Info() *ErrInfo {
	return e.info
}

// check interfaces
var (
	_ ProtoErr = (*CommandError)(nil)
)
