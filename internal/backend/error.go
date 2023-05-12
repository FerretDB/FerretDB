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

package backend

import (
	"errors"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
)

//go:generate ../../bin/stringer -linecomment -type ErrorCode

type ErrorCode int

const (
	_ ErrorCode = iota

	ErrCollectionDoesNotExist  // collection does not exist
	ErrCollectionAlreadyExists // collection already exists
)

type Error struct {
	Code ErrorCode
	err  error
}

func (err *Error) Error() string {
	return err.err.Error()
}

func (err *Error) Unwrap() error {
	return err.err
}

func checkError(err error, codes ...ErrorCode) {
	if !debugbuild.Enabled {
		return
	}

	if err == nil {
		return
	}

	var e *Error
	if !errors.As(err, &e) {
		panic(err)
	}

	if e.Code == 0 {
		panic(err)
	}

	if !slices.Contains(codes, e.Code) {
		panic(err)
	}
}

// check interfaces
var (
	_ error = (*Error)(nil)
)
