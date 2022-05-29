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

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFreeSpacingParse(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		input    string
		expected string
	}{
		"EmptyExpr": {``, ""},
		"MultilineExpr": {"(?=\t\t # Start lookahead\n\t\\" +
			"D*\t # non-digits\n\t\\d\t # one digit\n)\n\n## matching\n\\w*\t\t#word chars\n[A-Z] \t# one upper-case\n\\" +
			"w*# word chars\n$\t\t# end of string\n", "(?=\\D*\\d)\\w*[A-Z]\\w*$"},
		"WhitespaceEscapes": {`a\ b[ ]c`, `a\ b[ ]c`},
		"SpaceEscapeChar":   {`\ d`, `\ d`},
		"Quantifier":        {"o{1 0}", "o\\{10}"},
		//"SpaceInToken":         {`(A)\1 2`, `(A)\1 2`},
		//"SpaceInCurlyBrackets": {`\p{1 2}`, `\p{1 2}`},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.expected, freeSpacingParse(tc.input))
		})
	}
}

func TestIsQuantifier(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		input    string
		expected bool
	}{
		"Digits":               {"1532}", true},
		"ContentAfterBrackets": {"1532}4a,,{}", true},
		"Range":                {"12,33}", true},
		"EmptyInput":           {"", false},
		"EmptyBrackets":        {"}", false},
		"NonDigits":            {"12sd}", false},
		"Space":                {"4, 3}", false},
		"MultipleCommas":       {"1,2,3}", false},
		"EmptyBeforeComma":     {",2}", false},
		"EmptyAfterComma":      {"1,}", false},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.expected, isQuantifier(tc.input))
		})
	}
}
