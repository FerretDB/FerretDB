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

// Package iface provides converters from functions/methods to interfaces,
// similarly to http.HandlerFunc.
//
// This package is a workaround for https://github.com/golang/go/issues/47487.
package iface

import "fmt"

// stringer implements [fmt.Stringer].
type stringer struct {
	f func() string
}

// String implements [fmt.Stringer].
func (s stringer) String() string {
	return s.f()
}

// Stringer converts a function to [fmt.Stringer].
//
// It may be used to avoid adding the String method to the type that might be problematic.
func Stringer(f func() string) fmt.Stringer {
	return stringer{f: f}
}
