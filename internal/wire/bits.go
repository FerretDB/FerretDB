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

package wire

import "strings"

type flagBit uint32

type flags uint32

// FlagsSize represents flags size in bytes.
const FlagsSize = 4

func (flags flags) strings(bitStringer func(flagBit) string) []string {
	res := make([]string, 0, 2)
	for shift := 0; shift < 32; shift++ {
		bit := flags >> shift
		if bit&1 == 1 {
			res = append(res, bitStringer(1<<shift))
		}
	}
	return res
}

func (flags flags) string(bitStringer func(flagBit) string) string {
	res := flags.strings(bitStringer)
	return "[" + strings.Join(res, "|") + "]"
}
