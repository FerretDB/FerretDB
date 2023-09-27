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

// Package testdata provides vet tool test data.
package testdata

func testCorrect() {
	// below issue url will be changed to https://api.github.com/repos/FerretDB/FerretDB/issues/2733
	// and checked issue open or closed.
	// TODO https://github.com/FerretDB/FerretDB/issues/2733
}

func testCorrectForNow() {
	// TODO no URL
}

func testIncorrect() {
	// TODO: https://github.com/FerretDB/FerretDB/issues/2733 // want "invalid TODO comment"
}
