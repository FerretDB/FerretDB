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

package password

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// plainTestCase represents a test case for PLAIN authentication.
//
//nolint:vet // for readability
type plainTestCase struct {
	params   plainParams
	password string
	salt     []byte
	hash     []byte
}

// https://github.com/brycx/Test-Vector-Generation/blob/master/PBKDF2/pbkdf2-hmac-sha2-test-vectors.md
//
// Test Case 4 is skipped because it takes too long.
var plainTestCases = []plainTestCase{{
	params: plainParams{
		iterationCount: 1,
		saltLen:        4,
		hashLen:        20,
	},
	password: "password",
	salt:     []byte("salt"),
	hash:     must.NotFail(hex.DecodeString("120fb6cffcf8b32c43e7225256c4f837a86548c9")),
}, {
	params: plainParams{
		iterationCount: 2,
		saltLen:        4,
		hashLen:        20,
	},
	password: "password",
	salt:     []byte("salt"),
	hash:     must.NotFail(hex.DecodeString("ae4d0c95af6b46d32d0adff928f06dd02a303f8e")),
}, {
	params: plainParams{
		iterationCount: 4096,
		saltLen:        4,
		hashLen:        20,
	},
	password: "password",
	salt:     []byte("salt"),
	hash:     must.NotFail(hex.DecodeString("c5e478d59288c841aa530db6845c4c8d962893a0")),
}, {
	params: plainParams{
		iterationCount: 4096,
		saltLen:        36,
		hashLen:        25,
	},
	password: "passwordPASSWORDpassword",
	salt:     []byte("saltSALTsaltSALTsaltSALTsaltSALTsalt"),
	hash:     must.NotFail(hex.DecodeString("348c89dbcbd32b2f32d814b8116e84cf2b17347ebc1800181c")),
}, {
	params: plainParams{
		iterationCount: 4096,
		saltLen:        5,
		hashLen:        16,
	},
	password: "pass\x00word",
	salt:     []byte("sa\x00lt"),
	hash:     must.NotFail(hex.DecodeString("89b69d0516f829893c696226650a8687")),
}, {
	params: plainParams{
		iterationCount: 1,
		saltLen:        4,
		hashLen:        128,
	},
	password: "passwd",
	salt:     []byte("salt"),
	hash:     must.NotFail(hex.DecodeString("55ac046e56e3089fec1691c22544b605f94185216dde0465e68b9d57c20dacbc49ca9cccf179b645991664b39d77ef317c71b845b1e30bd509112041d3a19783c294e850150390e1160c34d62e9665d659ae49d314510fc98274cc79681968104b8f89237e69b2d549111868658be62f59bd715cac44a1147ed5317c9bae6b2a")), //nolint:lll // follows test vectors
}, {
	params: plainParams{
		iterationCount: 80000,
		saltLen:        4,
		hashLen:        128,
	},
	password: "Password",
	salt:     []byte("NaCl"),
	hash:     must.NotFail(hex.DecodeString("4ddcd8f60b98be21830cee5ef22701f9641a4418d04c0414aeff08876b34ab56a1d425a1225833549adb841b51c9b3176a272bdebba1d078478f62b397f33c8d62aae85a11cdde829d89cb6ffd1ab0e63a981f8747d2f2f9fe5874165c83c168d2eed1d2d5ca4052dec2be5715623da019b8c0ec87dc36aa751c38f9893d15c3")), //nolint:lll // follows test vectors
}, {
	params: plainParams{
		iterationCount: 4096,
		saltLen:        5,
		hashLen:        256,
	},
	password: "Password",
	salt:     []byte("sa\x00lt"),
	hash:     must.NotFail(hex.DecodeString("436c82c6af9010bb0fdb274791934ac7dee21745dd11fb57bb90112ab187c495ad82df776ad7cefb606f34fedca59baa5922a57f3e91bc0e11960da7ec87ed0471b456a0808b60dff757b7d313d4068bf8d337a99caede24f3248f87d1bf16892b70b076a07dd163a8a09db788ae34300ff2f2d0a92c9e678186183622a636f4cbce15680dfea46f6d224e51c299d4946aa2471133a649288eef3e4227b609cf203dba65e9fa69e63d35b6ff435ff51664cbd6773d72ebc341d239f0084b004388d6afa504eee6719a7ae1bb9daf6b7628d851fab335f1d13948e8ee6f7ab033a32df447f8d0950809a70066605d6960847ed436fa52cdfbcf261b44d2a87061")), //nolint:lll // follows test vectors
}}

func TestPlain(t *testing.T) {
	for i, tc := range plainTestCases {
		i, tc := i, tc
		t.Run(fmt.Sprintf("%d-%x", i, tc.hash), func(t *testing.T) {
			t.Parallel()

			require.Len(t, tc.hash, tc.params.hashLen, "invalid hash length")

			doc, hash, err := plainHashParams(tc.password, tc.salt, &tc.params)
			require.NoError(t, err)
			assert.Equal(t, tc.hash, hash, "hash mismatch")

			params, err := plainVerifyParams(tc.password, doc)
			require.NoError(t, err)
			require.NotNil(t, params, "params is nil")
			assert.Equal(t, tc.params, *params, "params mismatch")
		})
	}

	t.Run("Exported", func(t *testing.T) {
		doc1, err := PlainHash("password")
		require.NoError(t, err)

		doc2, err := PlainHash("password")
		require.NoError(t, err)

		testutil.AssertNotEqual(t, doc1, doc2)

		err = PlainVerify("password", doc1)
		assert.NoError(t, err)

		err = PlainVerify("password", doc2)
		assert.NoError(t, err)

		err = PlainVerify("wrong", doc2)
		assert.Error(t, err)
	})
}

func BenchmarkPlain(b *testing.B) {
	var err error

	b.Run("Exported", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err = PlainHash("password")
		}
	})

	require.NoError(b, err)
}
