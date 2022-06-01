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

// Package lazyerrors provides temporary error wrapping for lazy developers.
package lazyerrors

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type withStack struct {
	error
	pc uintptr
}

func (e withStack) Error() string {
	if e.pc == 0 {
		return e.error.Error()
	}

	f := frame(e.pc)
	if f.File == "" {
		return "[unknown] " + e.error.Error()
	}

	_, file := filepath.Split(f.File)
	l := file + ":" + strconv.Itoa(f.Line)
	if f.Function != "" {
		i := strings.LastIndex(f.Function, "/")
		l += " " + f.Function[i+1:]
	}

	return fmt.Sprintf("[%s] %s", l, e.error)
}

func (e withStack) Unwrap() error {
	return e.error
}

// New returns new error based on string, enriched with callers.
func New(s string) error {
	return withStack{
		error: errors.New(s),
		pc:    pc(),
	}
}

// Error returns new error based on err and ensures err is not nil.
func Error(err error) error {
	if err == nil {
		panic("err is nil")
	}

	return withStack{
		error: err,
		pc:    pc(),
	}
}

// Errorf returns formatted error enriched with callers.
func Errorf(format string, a ...any) error {
	return withStack{
		error: fmt.Errorf(format, a...),
		pc:    pc(),
	}
}
