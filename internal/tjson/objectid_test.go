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

package tjson

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/types"
)

var objectIDTestCases = []testCase{{
	name: "normal",
	v:    types.ObjectID{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
	j:    `{"$o":"010101010101010101010101"}`,
	s:    objectSchema,
}, {
	name: "EOF",
	j:    `{`,
	jErr: `unexpected EOF`,
	s:    objectSchema,
}}

func TestObjectID(t *testing.T) {
	t.Parallel()
	testJSON(t, objectIDTestCases, func() tjsontype { return new(objectIDType) })
}

func FuzzObjectID(f *testing.F) {
	fuzzJSON(f, objectIDTestCases, func() tjsontype { return new(objectIDType) })
}

func BenchmarkObjectID(b *testing.B) {
	benchmark(b, objectIDTestCases, func() tjsontype { return new(objectIDType) })
}
