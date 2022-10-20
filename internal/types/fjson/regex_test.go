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

package fjson

import (
	"testing"

	"github.com/AlekSi/pointer"
)

var regexTestCases = []testCase{{
	name: "normal",
	v:    pointer.To(regexType{Pattern: "hoffman", Options: "i"}),
	j:    `{"$r":"hoffman","o":"i"}`,
}, {
	name: "empty",
	v:    pointer.To(regexType{Pattern: "", Options: ""}),
	j:    `{"$r":"","o":""}`,
}}

func TestRegex(t *testing.T) {
	t.Parallel()
	testJSON(t, regexTestCases, func() fjsontype { return new(regexType) })
}
