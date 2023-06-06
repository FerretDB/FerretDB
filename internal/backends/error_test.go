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

package backends

import (
	"io"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorNormal(t *testing.T) {
	t.Parallel()

	pe := &fs.PathError{
		Op:   "open",
		Path: "database.db",
		Err:  io.EOF,
	}
	err := NewError(ErrorCodeCollectionDoesNotExist, pe)

	assert.NotErrorIs(t, err, pe, "internal error should be hidden")
	assert.NotErrorIs(t, err, io.EOF, "internal error should be hidden")

	var e *Error
	assert.ErrorAs(t, err, &e)
	assert.Equal(t, ErrorCodeCollectionDoesNotExist, e.code)
	assert.Equal(t, pe, e.err)

	assert.Equal(t, `ErrorCodeCollectionDoesNotExist: open database.db: EOF`, err.Error())
}

func TestErrorNil(t *testing.T) {
	t.Parallel()

	err := NewError(ErrorCodeCollectionDoesNotExist, nil)

	var e *Error
	assert.ErrorAs(t, err, &e)
	assert.Equal(t, ErrorCodeCollectionDoesNotExist, e.code)
	assert.Nil(t, e.err)

	assert.Equal(t, `ErrorCodeCollectionDoesNotExist: <nil>`, err.Error())
}
