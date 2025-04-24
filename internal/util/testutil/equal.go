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

package testutil

import (
	"testing"

	"github.com/FerretDB/wire/wirebson"
	"github.com/FerretDB/wire/wiretest"
)

// AssertEqual asserts that two BSON values are equal.
//
// Deprecated: use [wiretest.AssertEqual] instead.
func AssertEqual[T wirebson.Type](tb testing.TB, expected, actual T) bool {
	tb.Helper()

	return wiretest.AssertEqual(tb, expected, actual)
}
