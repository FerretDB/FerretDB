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

package mongoerrors

import (
	"fmt"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// Error represents MongoDB command error.
type Error struct {
	// Command's argument, operator, or aggregation pipeline stage that caused an error.
	// Used for metrics and telemetry.
	Argument string

	mongo.CommandError
}

// New creates a new Error.
//
// Existing errors should be converted using [Make], not created by this function.
func New(code Code, msg string) *Error {
	if code <= 0 {
		panic(fmt.Sprintf("invalid error code: %d", code))
	}

	return &Error{
		CommandError: mongo.CommandError{
			Code:    int32(code),
			Message: msg,
			Name:    code.String(),
		},
	}
}

// NewWithArgument creates a new Error caused by the given argument.
//
// Existing errors should be converted using [Make], not created by this function.
func NewWithArgument(code Code, msg, argument string) *Error {
	if code <= 0 {
		panic(fmt.Sprintf("invalid error code: %d", code))
	}

	return &Error{
		Argument: argument,
		CommandError: mongo.CommandError{
			Code:    int32(code),
			Message: msg,
			Name:    code.String(),
		},
	}
}

// Error implements error interface.
//
// We overload [mongo.CommandError]'s method to ensure that Error is always passed by pointer.
func (e *Error) Error() string {
	return fmt.Sprintf("%[1]s (%[1]d): %[2]v", Code(e.Code), e.Message)
}

// GoString implements fmt.GoStringer interface.
//
// It exists so %#v fmt verb could correctly print wrapped errors.
func (e *Error) GoString() string {
	return fmt.Sprintf(
		`&mongoerrors.Error{Code: %d, Name: %#q, Message: %#q, Argument: %#q, Wrapped: %s}`,
		e.Code, e.Name, e.Message, e.Argument, goString(e.Wrapped),
	)
}

// Msg returns this error as a OP_MSG message.
func (e *Error) Msg() *wire.OpMsg {
	return wire.MustOpMsg(
		"ok", float64(0),
		"errmsg", e.Message,
		"code", int32(e.Code),
		"codeName", e.Name,
	)
}

// Reply returns this error as a OP_REPLY message.
func (e *Error) Reply() *wire.OpReply {
	return must.NotFail(wire.NewOpReply(wirebson.MustDocument(
		"ok", float64(0),
		"errmsg", e.Message,
		"code", int32(e.Code),
		"codeName", e.Name,
	)))
}
