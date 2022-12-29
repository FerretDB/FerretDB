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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// WriteErrors represents a slice of protocol write errors.
// It could be returned for Update, Insert, Delete, and Replace operations.
type WriteErrors struct {
	errs []writeError
}

// NewWriteErrorMsg creates a new protocol write error with given ErrorCode and message.
func NewWriteErrorMsg(code ErrorCode, msg string) error {
	return &WriteErrors{
		errs: []writeError{{
			code: code,
			err:  msg,
		}},
	}
}

// Error implements error interface.
func (we *WriteErrors) Error() string {
	var err string

	for i, e := range we.errs {
		if i != 0 {
			err += ", "
		}

		err += e.err
	}

	return err
}

// Unwrap implements ProtoErr interface.
func (we *WriteErrors) Unwrap() error {
	for _, e := range we.errs {
		return errors.New(e.err)
	}

	return nil
}

// Code implements ProtoErr interface.
func (we *WriteErrors) Code() ErrorCode {
	for _, e := range we.errs {
		return e.code
	}

	return errUnset
}

// Document implements ProtoErr interface.
func (we *WriteErrors) Document() *types.Document {
	errs := must.NotFail(types.NewArray())

	for _, e := range we.errs {
		doc := must.NotFail(types.NewDocument())

		if e.index != nil {
			doc.Set("index", *e.index)
		}

		// Fields "code" and "errmsg" must always be filled in so that clients can parse the error message.
		// Otherwise, the mongo client would parse it as a CommandError.
		doc.Set("code", int32(e.code))
		doc.Set("errmsg", e.err)

		must.NoError(errs.Append(doc))
	}

	// "writeErrors" field must be present in the result document so that clients can parse it as WriteErrors.
	return must.NotFail(types.NewDocument(
		"ok", float64(1),
		"writeErrors", errs,
	))
}

// Info implements ProtoErr interface.
func (we *WriteErrors) Info() *ErrInfo {
	return nil
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
		we.errs = append(we.errs, *writeErr)

		return

	case errors.As(err, &cmdErr):
		we.errs = append(we.errs, writeError{err: cmdErr.Unwrap().Error(), code: cmdErr.code, index: &index})

		return
	}

	we.errs = append(we.errs, writeError{err: err.Error(), code: errInternalError, index: &index})
}

// Len returns the number of errors.
func (we *WriteErrors) Len() int {
	return len(we.errs)
}

// Merge merges the given WriteErrors with the current one and sets the given index.
func (we *WriteErrors) Merge(we2 *WriteErrors, index int32) {
	for _, e := range we2.errs {
		e.index = &index
		we.errs = append(we.errs, e)
	}
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

// check interfaces
var (
	_ ProtoErr = (*WriteErrors)(nil)
)
