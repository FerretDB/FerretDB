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

package bson

import (
	"testing"

	"github.com/AlekSi/pointer"
)

var boolTestCases = []testCase{{
	name: "false",
	v:    pointer.To(boolType(false)),
	b:    []byte{0x00},
}, {
	name: "true",
	v:    pointer.To(boolType(true)),
	b:    []byte{0x01},
}}

func TestBool(t *testing.T) {
	t.Parallel()
	testBinary(t, boolTestCases, func() bsontype { return new(boolType) })
}

func FuzzBool(f *testing.F) {
	fuzzBinary(f, boolTestCases, func() bsontype { return new(boolType) })
}

func BenchmarkBool(b *testing.B) {
	benchmark(b, boolTestCases, func() bsontype { return new(boolType) })
}
