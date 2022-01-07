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

package pg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtf8LocaleValidity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Case     string
		Expected bool
	}{
		{"en_US.utf8", true},
		{"en_US.utf-8", true},
		{"en_US.UTF8", true},
		{"en_US.UTF-8", true},
		{"en_UK.UTF-8", false},
		{"en_UK.utf--8", false},
		{"en_US", false},
		{"utf8", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Case, func(t *testing.T) {
			t.Parallel()

			actual := validUtf8Locale(tc.Case)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
