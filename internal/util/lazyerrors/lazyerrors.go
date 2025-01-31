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

// Package lazyerrors provides error wrapping with file location.
//
// Only one file location is captures for each error, not a full stack.
// If the chain is needed, don't forget to add links manually.
package lazyerrors

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// pc returns a program counter of the caller.
func pc() uintptr {
	pc := make([]uintptr, 1)
	runtime.Callers(3, pc)

	return pc[0]
}

// lazyerror wraps an error with a file location.
type lazyerror struct {
	error
	pc uintptr
}

// Error return the wrapped error message prefixed with file location.
func (le lazyerror) Error() string {
	if le.pc == 0 {
		return le.error.Error()
	}

	f, _ := runtime.CallersFrames([]uintptr{le.pc}).Next()
	if f.File == "" {
		return "[unknown] " + le.error.Error()
	}

	_, file := filepath.Split(f.File)
	l := file + ":" + strconv.Itoa(f.Line)
	if f.Function != "" {
		i := strings.LastIndex(f.Function, "/")
		l += " " + f.Function[i+1:]
	}

	return fmt.Sprintf("[%s] %s", l, le.error)
}

// GoString implements fmt.GoStringer interface.
//
// It exists so %#v fmt verb could correctly print wrapped errors.
func (le lazyerror) GoString() string {
	return fmt.Sprintf("lazyerror(%s)", le.Error())
}

// Unwrap returns the wrapped error.
func (le lazyerror) Unwrap() error {
	return le.error
}

// New returns new error with a given error string and file location.
func New(s string) error {
	return lazyerror{
		error: errors.New(s),
		pc:    pc(),
	}
}

// Error returns new error with a given non-nil error and file location.
func Error(err error) error {
	if err == nil {
		panic("err is nil")
	}

	return lazyerror{
		error: err,
		pc:    pc(),
	}
}

// Errorf returns new error with a given format string and file location.
func Errorf(format string, a ...any) error {
	return lazyerror{
		error: fmt.Errorf(format, a...),
		pc:    pc(),
	}
}
