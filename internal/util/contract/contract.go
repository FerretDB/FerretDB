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

// Package contract provides Design by Contract functionality.
package contract

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
)

// EnsureError checks that error is either nil or one of expected errors.
//
// If that is not the case, EnsureError panics in debug builds.
// It does nothing in non-debug builds.
func EnsureError(err error, expected ...error) {
	if !debugbuild.Enabled {
		return
	}

	if err == nil {
		return
	}

	if reflect.ValueOf(err).IsZero() {
		panic(fmt.Sprintf("EnsureError: invalid actual value %#v", err))
	}

	for _, target := range expected {
		if target == nil {
			panic(fmt.Sprintf("EnsureError: invalid expected value %#v", target))
		}

		if reflect.ValueOf(target).IsZero() {
			panic(fmt.Sprintf("EnsureError: invalid expected value %#v", target))
		}

		if errors.Is(err, target) {
			return
		}
	}

	panic(fmt.Sprintf("EnsureError: %#v", err))
}
