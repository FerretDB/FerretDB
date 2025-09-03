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

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Error represents MongoDB command error.
type Error struct {
	// Command's argument name, operator name, or aggregation pipeline stage name that caused an error.
	// Used for metrics and telemetry.
	//
	// It should not be the command name itself (e.g. "create" or "collStats"),
	// except when it is a problem with the collection name (i.e. create's or collStats's value).
	// It also should not be a user-supplied value (e.g. collection name or index name).
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
//
// We probably should remove that method and un-embed mongo.CommandError to avoid typed nil error confusion.
// TODO https://github.com/FerretDB/FerretDB/issues/4965
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
