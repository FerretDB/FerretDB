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
	"fmt"
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
	username string
	password string
	salt     []byte

	want *types.Document
	err  string
}

var scramSHA256TestCases = []scramSHA256TestCase{
	{
		params: scramSHA256Params{
			iterationCount: 1,
			saltLen:        4,
		},
		username: "username",
		password: "password",
		salt:     []byte("salt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "tWgTq9QWqLI2SkBpZeZSmGl7RzeuMuU3vWYYEpOFTvk=",
			"iterationCount", int32(1),
			"salt", "c2FsdA==",
			"serverKey", "cLmYEp4e6nRZDv4vrrpjYSt/FPP/Ekt/XVZVoDlrByw=",
		)),
	},
	{
		params: scramSHA256Params{
			iterationCount: 2,
			saltLen:        4,
		},
		username: "username",
		password: "password",
		salt:     []byte("salt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "db2Cdby2HHY1enQpujvJPfRJNlLyQ95MIEMwybJdFcI=",
			"iterationCount", int32(2),
			"salt", "c2FsdA==",
			"serverKey", "a0OWicFaTNUVr7ZJDEnGc0sn9GLSAUyannq6uYeSJRs=",
		)),
	},
	{
		params: scramSHA256Params{
			iterationCount: 4096,
			saltLen:        4,
		},
		username: "username",
		password: "password",
		salt:     []byte("salt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "lF4cRm/Jky763CN4HtxdHnjV4Q8AWTNlKvGmEFFU8IQ=",
			"iterationCount", int32(4096),
			"salt", "c2FsdA==",
			"serverKey", "ub8OgRsftnk2ccDMOt7ffHXNcikRkQkq1lh4xaAqrSw=",
		)),
	},
	{
		params: scramSHA256Params{
			iterationCount: 4096,
			saltLen:        36,
		},
		username: "username",
		password: "passwordPASSWORDpassword",
		salt:     []byte("saltSALTsaltSALTsaltSALTsaltSALTsalt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "kl1yVUP4s3BJMFrtaC4zLJycbv6k5yMBhVgocYmsYsU=",
			"iterationCount", int32(4096),
			"salt", "c2FsdFNBTFRzYWx0U0FMVHNhbHRTQUxUc2FsdFNBTFRzYWx0",
			"serverKey", "4QTMss7Dzi+pk8C8cyql++OaWqI0y/FyhXJI7W9acHI=",
		)),
	},
	{
		params: scramSHA256Params{
			iterationCount: 4096,
			saltLen:        5,
		},
		username: "username",
		password: "pass\x00word",
		salt:     []byte("sa\x00lt"),
		err:      "prohibited character",
	},
	{
		params: scramSHA256Params{
			iterationCount: 1,
			saltLen:        4,
		},
		username: "username",
		password: "passwd",
		salt:     []byte("salt"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "dcwmgrDYICpRpKjwHLKxZ/21/go62U106s5V4i9v+Q8=",
			"iterationCount", int32(1),
			"salt", "c2FsdA==",
			"serverKey", "qFes4m5Z84MaC2hSJqCR2e/FBz7goMVu/RTRNnb5Fj0=",
		)),
	},
	{
		params: scramSHA256Params{
			iterationCount: 80000,
			saltLen:        4,
		},
		username: "username",
		password: "Password",
		salt:     []byte("NaCl"),
		want: must.NotFail(types.NewDocument(
			"storedKey", "EI8wmB+eWKGZ8k+dln75YQnX8lj+MBoG6+eH9AzR1e4=",
			"iterationCount", int32(80000),
			"salt", "TmFDbA==",
			"serverKey", "Wbwo6JsaJrZ/1Bf7F+45jY2VURuezLXxADxUuzWdZ/4=",
		)),
	},
	{
		params: scramSHA256Params{
			iterationCount: 4096,
			saltLen:        5,
		},
		username: "username",
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
	for i, tc := range scramSHA256TestCases {
		i, tc := i, tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

			doc, err := scramSHA256HashParams(tc.username, tc.password, tc.salt, &tc.params)

			if tc.err != "" {
				assert.Contains(t, err.Error(), tc.err)
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
			must.NoError(err)

			// Check if the generated authentication is valid by simulating a conversation.
			conv := scramServer.NewConversation()

			client, err := scram.SHA256.NewClient(tc.username, tc.password, "")
			require.NoError(t, err)

			resp, err := client.NewConversation().Step("")
			require.NoError(t, err)

			resp, err = conv.Step(resp)
			require.NoError(t, err)
			assert.NotEmpty(t, resp)

			_, err = conv.Step("wrong")
			assert.NotNil(t, err)
		})
	}

	t.Run("Exported", func(t *testing.T) {
		doc1, err := SCRAMSHA256Hash("username", "password")
		require.NoError(t, err)

		doc2, err := SCRAMSHA256Hash("username", "password")
		require.NoError(t, err)

		testutil.AssertNotEqual(t, doc1, doc2)
	})
}

func BenchmarkSCRAMSHA256(b *testing.B) {
	var err error

	b.Run("Exported", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, err = SCRAMSHA256Hash("username", "password")
		}
	})

	require.NoError(b, err)
}
