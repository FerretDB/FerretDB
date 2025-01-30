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

package scram

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

// parseMessageTestCase represents a test case for [parseMessage].
//
//nolint:vet // for readability
type parseMessageTestCase struct {
	name         string
	msg          string
	expected     *message
	expectedType func(*message) bool
}

// parseMessageTestCases is a set of test cases for [parseMessage].
var parseMessageTestCases = []parseMessageTestCase{
	{
		name: `rfc5802-example-client-first`,
		msg:  `n,,n=user,r=fyko+d2lbbFgONRv9qkxdawL`,
		expected: &message{
			gs2: "n,",
			n:   "user",
			r:   "fyko+d2lbbFgONRv9qkxdawL",
		},
		expectedType: (*message).isClientFirst,
	},
	{
		name: `rfc5802-example-server-first`,
		msg:  `r=fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j,s=QSXCR+Q6sek8bf92,i=4096`,
		expected: &message{
			r: "fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j",
			s: "QSXCR+Q6sek8bf92",
			i: 4096,
		},
		expectedType: (*message).isServerFirst,
	},
	{
		name: `rfc5802-example-client-final`,
		msg:  `c=biws,r=fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j,p=v0X8v3Bz2T0CJGbJQyF0X+HI4Ts=`,
		expected: &message{
			c: "biws",
			r: "fyko+d2lbbFgONRv9qkxdawL3rfcNHYJY1ZVvWVs7j",
			p: "v0X8v3Bz2T0CJGbJQyF0X+HI4Ts=",
		},
		expectedType: (*message).isClientFinal,
	},
	{
		name: `rfc5802-example-server-final`,
		msg:  `v=rmF9pqV8S7suAoZWja4dJRkFsKQ=`,
		expected: &message{
			v: "rmF9pqV8S7suAoZWja4dJRkFsKQ=",
		},
		expectedType: (*message).isServerFinal,
	},
}

func TestParseMessage(t *testing.T) {
	t.Parallel()

	for _, tc := range parseMessageTestCases {
		require.NotEmpty(t, tc.name)

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.expected != nil {
				require.Equal(t, tc.msg, tc.expected.String())
			}

			actual, err := parseMessage(tc.msg, testutil.Logger(t))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.msg, actual.String())

			require.True(t, tc.expectedType(actual))

			var found bool

			for _, f := range []func(*message) bool{
				(*message).isClientFirst,
				(*message).isServerFirst,
				(*message).isClientFinal,
				(*message).isServerFinal,
			} {
				ok := f(actual)
				if ok {
					require.False(t, found, "multiple functions returned true")
					found = true
				}
			}
		})
	}
}

func FuzzParseMessage(f *testing.F) {
	for _, tc := range parseMessageTestCases {
		f.Add(tc.msg)
	}

	f.Fuzz(func(t *testing.T, msg string) {
		t.Parallel()

		actual, err := parseMessage(msg, testutil.Logger(t))
		if err != nil {
			assert.Nil(t, actual)
			return
		}

		assert.Equal(t, msg, actual.String())

		var found bool

		for _, f := range []func(*message) bool{
			(*message).isClientFirst,
			(*message).isServerFirst,
			(*message).isClientFinal,
			(*message).isServerFinal,
		} {
			ok := f(actual)
			if ok {
				require.False(t, found, "multiple functions returned true")
				found = true
			}
		}
	})
}
