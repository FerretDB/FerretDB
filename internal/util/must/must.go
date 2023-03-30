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

// Package must provides helper functions that panic on error.
package must

// NotFail panics if the error is not nil, returns res otherwise.
//
// Use that function only for static initialization, test code, or code that "can't" fail.
// When in doubt, don't.
func NotFail[T any](res T, err error) T {
	if err != nil {
		panic(err)
	}
	return res
}

// NoError panics if the error is not nil.
//
// Use that function only for static initialization, test code, or code that "can't" fail.
// When in doubt, don't.
func NoError(err error) {
	if err != nil {
		panic(err)
	}
}

// BeTrue panic if the b is not true.
//
// Use that function only for static initialization, test code, or statemants that
// "can't" be false. When in doubt, don't.
func BeTrue(b bool) {
	if !b {
		panic("not true")
	}
}
