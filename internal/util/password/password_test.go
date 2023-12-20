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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

type testCase struct {
	encoded string
	hash    []byte
	salt    []byte
	params  params
}

var testCases = map[string]testCase{
	// https://github.com/P-H-C/phc-winner-argon2/blob/f57e61e19229e23c4445b85494dbf7c07de721cb/src/test.c#L233-L264
	"argon2-1": {
		encoded: "$argon2id$v=19$m=65536,t=2,p=1$c29tZXNhbHQ$CTFhFdXPJO1aFaMaO6Mm5c8y7cJHAph8ArZWb2GRPPc",
		hash:    must.NotFail(hex.DecodeString("09316115d5cf24ed5a15a31a3ba326e5cf32edc24702987c02b6566f61913cf7")),
		salt:    []byte("somesalt"),
		params: params{
			m: 1 << 16,
			t: 2,
			p: 1,
		},
	},
	"argon2-2": {
		encoded: "$argon2id$v=19$m=262144,t=2,p=1$c29tZXNhbHQ$eP4eyR+zqlZX1y5xCFTkw9m5GYx0L5YWwvCFvtlbLow",
		hash:    must.NotFail(hex.DecodeString("78fe1ec91fb3aa5657d72e710854e4c3d9b9198c742f9616c2f085bed95b2e8c")),
		salt:    []byte("somesalt"),
		params: params{
			m: 1 << 18,
			t: 2,
			p: 1,
		},
	},
	"argon2-3": {
		encoded: "$argon2id$v=19$m=256,t=2,p=1$c29tZXNhbHQ$nf65EOgLrQMR/uIPnA4rEsF5h7TKyQwu9U1bMCHGi/4",
		hash:    must.NotFail(hex.DecodeString("9dfeb910e80bad0311fee20f9c0e2b12c17987b4cac90c2ef54d5b3021c68bfe")),
		salt:    []byte("somesalt"),
		params: params{
			m: 1 << 8,
			t: 2,
			p: 1,
		},
	},
	"argon2-4": {
		encoded: "$argon2id$v=19$m=256,t=2,p=2$c29tZXNhbHQ$bQk8UB/VmZZF4Oo79iDXuL5/0ttZwg2f/5U52iv1cDc",
		hash:    must.NotFail(hex.DecodeString("6d093c501fd5999645e0ea3bf620d7b8be7fd2db59c20d9fff9539da2bf57037")),
		salt:    []byte("somesalt"),
		params: params{
			m: 1 << 8,
			t: 2,
			p: 2,
		},
	},
	"argon2-5": {
		encoded: "$argon2id$v=19$m=65536,t=1,p=1$c29tZXNhbHQ$9qWtwbpyPd3vm1rB1GThgPzZ3/ydHL92zKL+15XZypg",
		hash:    must.NotFail(hex.DecodeString("f6a5adc1ba723dddef9b5ac1d464e180fcd9dffc9d1cbf76cca2fed795d9ca98")),
		salt:    []byte("somesalt"),
		params: params{
			m: 1 << 16,
			t: 1,
			p: 1,
		},
	},
	"argon2-6": {
		encoded: "$argon2id$v=19$m=65536,t=4,p=1$c29tZXNhbHQ$kCXUjmjvc5XMqQedpMTsOv+zyJEf5PhtGiUghW9jFyw",
		hash:    must.NotFail(hex.DecodeString("9025d48e68ef7395cca9079da4c4ec3affb3c8911fe4f86d1a2520856f63172c")),
		salt:    []byte("somesalt"),
		params: params{
			m: 1 << 16,
			t: 4,
			p: 1,
		},
	},
	"argon2-7": {
		encoded: "$argon2id$v=19$m=65536,t=2,p=1$c29tZXNhbHQ$C4TWUs9rDEvq7w3+J4umqA32aWKB1+DSiRuBfYxFj94",
		hash:    must.NotFail(hex.DecodeString("0b84d652cf6b0c4beaef0dfe278ba6a80df6696281d7e0d2891b817d8c458fde")),
		salt:    []byte("somesalt"),
		params: params{
			m: 1 << 16,
			t: 2,
			p: 1,
		},
	},
	"argon2-8": {
		encoded: "$argon2id$v=19$m=65536,t=2,p=1$ZGlmZnNhbHQ$vfMrBczELrFdWP0ZsfhWsRPaHppYdP3MVEMIVlqoFBw",
		hash:    must.NotFail(hex.DecodeString("bdf32b05ccc42eb15d58fd19b1f856b113da1e9a5874fdcc544308565aa8141c")),
		salt:    []byte("diffsalt"),
		params: params{
			m: 1 << 16,
			t: 2,
			p: 1,
		},
	},
}

func TestDecodeEncode(t *testing.T) {
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.Regexp(t, encodedRe, tc.encoded)

			hash, salt, params, err := decode(tc.encoded)
			require.NoError(t, err)
			assert.Equal(t, tc.hash, hash, "hash mismatch")
			assert.Equal(t, tc.salt, salt, "salt mismatch")
			assert.Equal(t, tc.params, params, "params mismatch")

			encoded := encode(hash, salt, params)
			assert.Equal(t, tc.encoded, encoded, "encoded mismatch")
		})
	}
}

func FuzzDecodeEncode(f *testing.F) {
	for _, tc := range testCases {
		f.Add(tc.encoded)
	}

	f.Fuzz(func(t *testing.T, encoded string) {
		hash, salt, params, err := decode(encoded)
		if err != nil {
			t.Skip()
		}

		encoded2 := encode(hash, salt, params)
		// encoded2 might be different from encoded due to base64 without padding

		hash2, salt2, params2, err := decode(encoded2)
		require.NoError(t, err)
		assert.Equal(t, hash, hash2, "hash mismatch")
		assert.Equal(t, salt, salt2, "salt mismatch")
		assert.Equal(t, params, params2, "params mismatch")

		res2 := encode(hash, salt, params)
		assert.Equal(t, encoded2, res2, "res mismatch")
	})
}
