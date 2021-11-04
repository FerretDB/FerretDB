// Copyright 2021 Baltoro OÜ.
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

package bson

import (
	"testing"
)

var regexTestCases = []testCase{
	// TODO add test cases
}

func TestRegex(t *testing.T) {
	t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		t.Parallel()
		testBinary(t, regexTestCases, func() bsontype { return new(Regex) })
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		testJSON(t, regexTestCases, func() bsontype { return new(Regex) })
	})
}

func FuzzRegexBinary(f *testing.F) {
	fuzzBinary(f, regexTestCases, func() bsontype { return new(Regex) })
}

func FuzzRegexJSON(f *testing.F) {
	fuzzJSON(f, regexTestCases, func() bsontype { return new(Regex) })
}

func BenchmarkRegex(b *testing.B) {
	benchmark(b, regexTestCases, func() bsontype { return new(Regex) })
}
