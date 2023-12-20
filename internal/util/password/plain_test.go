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

// plainTestCase represents a test case for PLAIN authentication using Argon2id.
type plainTestCase struct {
	params   plainParams
	password string
	salt     []byte
	hash     []byte
}

// Tests from the reference implementation at
// https://github.com/P-H-C/phc-winner-argon2/blob/f57e61e19229e23c4445b85494dbf7c07de721cb/src/test.c#L233-L264
//
// void hashtest(
// uint32_t version, uint32_t t, uint32_t m, uint32_t p, char *pwd,
// char *salt, char *hexref, char *mcfref, argon2_type type
// )
//
// m there is 2^m, which is the same as 1<<m.
var plainTestCases = []plainTestCase{{
	params: plainParams{
		t: 2,
		m: 1 << 16,
		p: 1,
	},
	password: "password",
	salt:     []byte("somesalt"),
	hash:     must.NotFail(hex.DecodeString("09316115d5cf24ed5a15a31a3ba326e5cf32edc24702987c02b6566f61913cf7")),
}, {
	params: plainParams{
		t: 2,
		m: 1 << 18,
		p: 1,
	},
	password: "password",
	salt:     []byte("somesalt"),
	hash:     must.NotFail(hex.DecodeString("78fe1ec91fb3aa5657d72e710854e4c3d9b9198c742f9616c2f085bed95b2e8c")),
}, {
	params: plainParams{
		t: 2,
		m: 1 << 8,
		p: 1,
	},
	password: "password",
	salt:     []byte("somesalt"),
	hash:     must.NotFail(hex.DecodeString("9dfeb910e80bad0311fee20f9c0e2b12c17987b4cac90c2ef54d5b3021c68bfe")),
}, {
	params: plainParams{
		t: 2,
		m: 1 << 8,
		p: 2,
	},
	password: "password",
	salt:     []byte("somesalt"),
	hash:     must.NotFail(hex.DecodeString("6d093c501fd5999645e0ea3bf620d7b8be7fd2db59c20d9fff9539da2bf57037")),
}, {
	params: plainParams{
		t: 1,
		m: 1 << 16,
		p: 1,
	},
	password: "password",
	salt:     []byte("somesalt"),
	hash:     must.NotFail(hex.DecodeString("f6a5adc1ba723dddef9b5ac1d464e180fcd9dffc9d1cbf76cca2fed795d9ca98")),
}, {
	params: plainParams{
		t: 4,
		m: 1 << 16,
		p: 1,
	},
	password: "password",
	salt:     []byte("somesalt"),
	hash:     must.NotFail(hex.DecodeString("9025d48e68ef7395cca9079da4c4ec3affb3c8911fe4f86d1a2520856f63172c")),
}, {
	params: plainParams{
		t: 2,
		m: 1 << 16,
		p: 1,
	},
	password: "differentpassword",
	salt:     []byte("somesalt"),
	hash:     must.NotFail(hex.DecodeString("0b84d652cf6b0c4beaef0dfe278ba6a80df6696281d7e0d2891b817d8c458fde")),
}, {
	params: plainParams{
		t: 2,
		m: 1 << 16,
		p: 1,
	},
	password: "password",
	salt:     []byte("diffsalt"),
	hash:     must.NotFail(hex.DecodeString("bdf32b05ccc42eb15d58fd19b1f856b113da1e9a5874fdcc544308565aa8141c")),
}}

func TestPlain(t *testing.T) {
	for i, tc := range plainTestCases {
		t.Run(fmt.Sprintf("%d-%x", i, tc.hash), func(t *testing.T) {
			tc.params.tagLen = uint32(len(tc.hash))
			tc.params.saltLen = len(tc.salt)

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

	for i, tc := range plainTestCases {
		b.Run(fmt.Sprintf("%d-%x", i, tc.hash), func(b *testing.B) {
			b.ReportAllocs()

			tc.params.tagLen = uint32(len(tc.hash))
			tc.params.saltLen = len(tc.salt)

			var hash []byte

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, hash, err = plainHashParams(tc.password, tc.salt, &tc.params)
			}

			b.StopTimer()

			require.NoError(b, err)
			assert.Equal(b, tc.hash, hash, "hash mismatch")
		})
	}

	b.Run("Exported", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err = PlainHash("password")
		}
	})

	require.NoError(b, err)
}
