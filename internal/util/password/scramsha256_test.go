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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// scramSHA256TestCase represents a test case for SCRAM-SHA-256 authentication.
//
//nolint:vet // for readability
type scramSHA256TestCase struct {
	params   scramSHA256Params
	password string
	salt     []byte

	want *types.Document
	err  string
}

// Test cases for the SCRAM-SHA-256 authentication.
var scramSHA256TestCases = map[string]scramSHA256TestCase{
	// Test vector generated with db.runCommand({createUser: "user", pwd: "pencil", roles: []})
	"FromMongoDB": {
		params: scramSHA256Params{
			iterationCount: 15000,
			saltLen:        28,
		},
		password: "pencil",
		salt:     must.NotFail(base64.StdEncoding.DecodeString("vXan6ZbWmm5i+f+mKY598rnIfoAGGp+G9NP0qQ==")),
		want: must.NotFail(types.NewDocument(
			"storedKey", "bNxFkKtMt93v+ha80yJsDG6Xes3GOMh5qsRzwkcF85s=",
			"iterationCount", int32(15000),
			"salt", "vXan6ZbWmm5i+f+mKY598rnIfoAGGp+G9NP0qQ==",
			"serverKey", "1m33jRKioBEVpJzDdJeG5SgKPEmhPNx3A0jS4fINVyQ=",
		)),
	},

	// Test vector generated with db.runCommand({createUser: "user", pwd: "password", roles: []})
	"FromMongoDB2": {
		params: scramSHA256Params{
			iterationCount: 15000,
			saltLen:        28,
		},
		password: "password",
		salt:     must.NotFail(base64.StdEncoding.DecodeString("4vbrJBkaleBWRqgdXri8Otu1pwLCoX5BCUoa1Q==")),
		want: must.NotFail(types.NewDocument(
			"storedKey", "1442RVPbzP5LhF3i/2Ld19Xj8TGfgK6XPy0KEbTL5so=",
			"iterationCount", int32(15000),
			"salt", "4vbrJBkaleBWRqgdXri8Otu1pwLCoX5BCUoa1Q==",
			"serverKey", "JEbgbKWzWtOJV5qHOXQL3pV5lzhFLzPEtC5wonu+HmU=",
		)),
	},

	"BadSaltLength": {
		params: scramSHA256Params{
			iterationCount: 15000,
			saltLen:        28,
		},
		password: "password",
		salt:     []byte("short"),
		err:      "unexpected salt length: 5",
	},
	"ProhibitedCharacter": {
		params: scramSHA256Params{
			iterationCount: 4096,
			saltLen:        5,
		},
		password: "pass\x00word",
		salt:     []byte("sa\x00lt"),
		err:      "prohibited character",
	},

	// The following checks were inspired by the test cases for the PLAIN method in plain_test.go
	// https://github.com/brycx/Test-Vector-Generation/blob/master/PBKDF2/pbkdf2-hmac-sha2-test-vectors.md
	"1Iteration": {
		params: scramSHA256Params{
			iterationCount: 1,
			saltLen:        4,
		},
		password: "password",
		salt:     []byte("salt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "tWgTq9QWqLI2SkBpZeZSmGl7RzeuMuU3vWYYEpOFTvk=",
			"iterationCount", int32(1),
			"salt", "c2FsdA==",
			"serverKey", "cLmYEp4e6nRZDv4vrrpjYSt/FPP/Ekt/XVZVoDlrByw=",
		)),
	},
	"2Iterations": {
		params: scramSHA256Params{
			iterationCount: 2,
			saltLen:        4,
		},
		password: "password",
		salt:     []byte("salt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "db2Cdby2HHY1enQpujvJPfRJNlLyQ95MIEMwybJdFcI=",
			"iterationCount", int32(2),
			"salt", "c2FsdA==",
			"serverKey", "a0OWicFaTNUVr7ZJDEnGc0sn9GLSAUyannq6uYeSJRs=",
		)),
	},
	"4096Iterations": {
		params: scramSHA256Params{
			iterationCount: 4096,
			saltLen:        4,
		},
		password: "password",
		salt:     []byte("salt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "lF4cRm/Jky763CN4HtxdHnjV4Q8AWTNlKvGmEFFU8IQ=",
			"iterationCount", int32(4096),
			"salt", "c2FsdA==",
			"serverKey", "ub8OgRsftnk2ccDMOt7ffHXNcikRkQkq1lh4xaAqrSw=",
		)),
	},
	"DifferentSalt": {
		params: scramSHA256Params{
			iterationCount: 4096,
			saltLen:        36,
		},
		password: "passwordPASSWORDpassword",
		salt:     []byte("saltSALTsaltSALTsaltSALTsaltSALTsalt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "kl1yVUP4s3BJMFrtaC4zLJycbv6k5yMBhVgocYmsYsU=",
			"iterationCount", int32(4096),
			"salt", "c2FsdFNBTFRzYWx0U0FMVHNhbHRTQUxUc2FsdFNBTFRzYWx0",
			"serverKey", "4QTMss7Dzi+pk8C8cyql++OaWqI0y/FyhXJI7W9acHI=",
		)),
	},
	"DifferentPassword": {
		params: scramSHA256Params{
			iterationCount: 1,
			saltLen:        4,
		},
		password: "passwd",
		salt:     []byte("salt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "dcwmgrDYICpRpKjwHLKxZ/21/go62U106s5V4i9v+Q8=",
			"iterationCount", int32(1),
			"salt", "c2FsdA==",
			"serverKey", "qFes4m5Z84MaC2hSJqCR2e/FBz7goMVu/RTRNnb5Fj0=",
		)),
	},
	"NaCl": {
		params: scramSHA256Params{
			iterationCount: 80000,
			saltLen:        4,
		},
		password: "Password",
		salt:     []byte("NaCl"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "EI8wmB+eWKGZ8k+dln75YQnX8lj+MBoG6+eH9AzR1e4=",
			"iterationCount", int32(80000),
			"salt", "TmFDbA==",
			"serverKey", "Wbwo6JsaJrZ/1Bf7F+45jY2VURuezLXxADxUuzWdZ/4=",
		)),
	},
	"00Salt": {
		params: scramSHA256Params{
			iterationCount: 4096,
			saltLen:        5,
		},
		password: "Password",
		salt:     []byte("sa\x00lt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "BHJbwIZ9YCvb+dFApByBLe4jR7gquvBm3kPApnWCylk=",
			"iterationCount", int32(4096),
			"salt", "c2EAbHQ=",
			"serverKey", "qpmMln7z50yoT66R1lT55pUvBxu8BqZU4X8dRp38NWo=",
		)),
	},
}

func TestSCRAMSHA256(t *testing.T) {
	t.Parallel()

	for name, tc := range scramSHA256TestCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			doc, err := scramSHA256HashParams(tc.password, tc.salt, &tc.params)

			if tc.err != "" {
				assert.ErrorContains(t, err, tc.err)
				return
			}

			require.NoError(t, err)
			testutil.AssertEqual(t, tc.want, doc)

			scramServer, err := scram.SHA256.NewServer(func(username string) (scram.StoredCredentials, error) {
				return scram.StoredCredentials{
					KeyFactors: scram.KeyFactors{
						Salt:  string(tc.salt),
						Iters: tc.params.iterationCount,
					},
					StoredKey: []byte(must.NotFail(doc.Get("storedKey")).(string)),
					ServerKey: []byte(must.NotFail(doc.Get("serverKey")).(string)),
				}, nil
			})
			require.NoError(t, err)

			// Check if the generated authentication is valid by simulating a conversation.
			conv := scramServer.NewConversation()

			client, err := scram.SHA256.NewClient("user", tc.password, "")
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

		doc1, err := SCRAMSHA256Hash("password")
		require.NoError(t, err)

		doc2, err := SCRAMSHA256Hash("password")
		require.NoError(t, err)

		testutil.AssertNotEqual(t, doc1, doc2)

		// salt is 30 bytes, but a value length increases ~33% when base64-encoded
		assert.Len(t, must.NotFail(doc1.Get("salt")), 40)
		assert.Len(t, must.NotFail(doc2.Get("salt")), 40)
	})
}

func BenchmarkSCRAMSHA256(b *testing.B) {
	var err error

	b.Run("Exported", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err = SCRAMSHA256Hash("password")
		}
	})

	require.NoError(b, err)
}
