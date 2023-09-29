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
	// TODO https://github.com/FerretDB/FerretDB/issues/2733 // want "valid TODO comment"
}

func testCorrectForNow() {
	// TODO no URL // want "invalid TODO comment because we need https"
}

func testIncorrect() {
	// TODO: https://github.com/FerretDB/FerretDB/issues/2733 // want "invalid TODO comment"
}

func testCorrectButClosedIssue() {
	// TODO: https://github.com/FerretDB/FerretDB/issues/3341 // want "invalid TODO comment because closed issue"
}

func testIncorrectCorrect() {
	// TODO: https://github.com/FerretDB/FerretDB/issues/2733 // want "valid TODO comment open issue"
}
