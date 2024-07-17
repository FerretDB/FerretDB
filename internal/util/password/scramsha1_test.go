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
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// scramSHA1TestCase represents a test case for SCRAM-SHA-1 authentication.
//
//nolint:vet // for readability
type scramSHA1TestCase struct {
	params   scramParams
	username string
	password Password
	salt     []byte

	want *bson.Document
	err  string
}

// Test cases for the SCRAM-SHA-1 authentication.
var scramSHA1TestCases = map[string]scramSHA1TestCase{
	// Test vector generated with db.runCommand({createUser: "user", pwd: "pencil", roles: []})
	"FromMongoDB": {
		params: scramParams{
			iterationCount: 10000,
			saltLen:        16,
		},
		username: "user",
		password: WrapPassword("pencil"),
		salt:     must.NotFail(base64.StdEncoding.DecodeString("55hyMPh69qfbfXPueAsr6g==")),
		want: must.NotFail(bson.NewDocument(
			"iterationCount", int32(10000),
			"salt", "55hyMPh69qfbfXPueAsr6g==",
			"storedKey", "W+jN9/MwzC8uhAIOOViZOUiSt14=",
			"serverKey", "XnkE60pg5UdvNe9nI9qF8VFV7Og=",
		)),
	},

	// Test vector generated with db.runCommand({createUser: "user", pwd: "password", roles: []})
	"FromMongoDB2": {
		params: scramParams{
			iterationCount: 10000,
			saltLen:        16,
		},
		username: "user",
		password: WrapPassword("password"),
		salt:     must.NotFail(base64.StdEncoding.DecodeString("OjPS7S2yaYBaJsRTCzahWQ==")),
		want: must.NotFail(bson.NewDocument(
			"iterationCount", int32(10000),
			"salt", "OjPS7S2yaYBaJsRTCzahWQ==",
			"storedKey", "fvtSnYGbBxKrXwbh4nAaUyiYMgg=",
			"serverKey", "TecAz+P5gUkeFwSB4QxzRZhzryc=",
		)),
	},

	"BadSaltLength": {
		params: scramParams{
			iterationCount: 15000,
			saltLen:        16,
		},
		password: WrapPassword("password"),
		salt:     []byte("short"),
		err:      "unexpected salt length: 5",
	},
}

func TestSCRAMSHA1Variation(t *testing.T) {
	t.Parallel()

	for name, tc := range scramSHA1TestCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doc, err := scramSHA1VariationHashParams(tc.username, tc.password, tc.salt, &tc.params)

			if tc.err != "" {
				assert.ErrorContains(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, doc)

			scramServer, err := scram.SHA1.NewServer(func(username string) (scram.StoredCredentials, error) {
				return scram.StoredCredentials{
					KeyFactors: scram.KeyFactors{
						Salt:  string(tc.salt),
						Iters: tc.params.iterationCount,
					},
					StoredKey: []byte(doc.Get("storedKey").(string)),
					ServerKey: []byte(doc.Get("serverKey").(string)),
				}, nil
			})
			require.NoError(t, err)

			// Check if the generated authentication is valid by simulating a conversation.
			conv := scramServer.NewConversation()

			client, err := scram.SHA1.NewClient(tc.username, tc.password.Password(), "")
			require.NoError(t, err)

			resp, err := client.NewConversation().Step("")
			require.NoError(t, err)

			resp, err = conv.Step(resp)
			require.NoError(t, err)
			assert.NotEmpty(t, resp)

			_, err = conv.Step("wrong")
			assert.Error(t, err)
		})
	}

	t.Run("Exported", func(t *testing.T) {
		t.Parallel()

		doc1, err := SCRAMSHA1VariationHash("user", WrapPassword("password"))
		require.NoError(t, err)

		doc2, err := SCRAMSHA1VariationHash("user", WrapPassword("password"))
		require.NoError(t, err)
		require.NotEqual(t, doc1, doc2)

		salt := doc1.Get("salt").(string)
		assert.Len(t, must.NotFail(base64.StdEncoding.DecodeString(salt)), 16)

		salt = doc2.Get("salt").(string)
		assert.Len(t, must.NotFail(base64.StdEncoding.DecodeString(salt)), 16)
	})
}

func BenchmarkSCRAMSHA1(b *testing.B) {
	var err error

	b.Run("Exported", func(b *testing.B) {
		b.ReportAllocs()

		for range b.N {
			_, err = SCRAMSHA1VariationHash("user", WrapPassword("password"))
		}
	})

	require.NoError(b, err)
}
