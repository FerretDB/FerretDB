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

package handler

import (
	"encoding/base64"
	"testing"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/stretchr/testify/assert"
	"github.com/xdg-go/scram"
)

// for user 'username:password'
const (
	salt      = "7jW5ZOczj05P4wyNc21OikIuSliPN9rw4sEoGQ=="
	storedKey = "F8hTLrnZscuuszfrh+4nupyjPA40cp+gfzy1Hsc3O3c="
	serverKey = "d4P+d81D31XHwvfQA3jwgTmkivZfXTD/nBASm77Dwv0="
)

func TestSaslStartSCRAM(t *testing.T) {
	validPayload := []byte("biwsbj11c2VybmFtZSxyPTRFSkNKcmVNejV1cDhnME5oa0E1L3UyTDdSRXVpOUFs")

	for name, tc := range map[string]struct { //nolint:vet // for readability
		doc *types.Document

		// expected results
		username string
		password string
		err      error
	}{
		"binaryPayload": {
			doc:      must.NotFail(types.NewDocument("payload", types.Binary{B: validPayload})),
			username: "username",
			password: "password",
			err:      nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			client, err := scram.SHA256.NewClient(tc.username, tc.password, "")
			assert.NoError(t, err)

			// log the crendentials for the user created by the mongo shell
			cf := scram.KeyFactors{
				Salt:  salt,
				Iters: 4096,
			}
			cred := client.GetStoredCredentials(cf)
			t.Log(base64.StdEncoding.EncodeToString(cred.StoredKey))
			t.Log(base64.StdEncoding.EncodeToString(cred.ServerKey))

			b, err := base64.StdEncoding.DecodeString(salt)
			assert.NoError(t, err)
			t.Log(b)

			b, err = base64.StdEncoding.DecodeString(storedKey)
			assert.NoError(t, err)
			t.Log(b)

			b, err = base64.StdEncoding.DecodeString(serverKey)
			assert.NoError(t, err)
			t.Log(b)
		})
	}
}
