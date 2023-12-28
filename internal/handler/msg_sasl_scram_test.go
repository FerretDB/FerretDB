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
	"testing"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/stretchr/testify/assert"
	"github.com/xdg-go/scram"
)

const (
	nonce = "rOprNGfwEbeRWgbNEkqO"
	salt  = "W22ZaJ0SNY7soEsUEjb6gQ"
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

			cf := scram.KeyFactors{
				Salt:  salt,
				Iters: 4096,
			}
			cred := client.GetStoredCredentials(cf)
			t.Log(string(cred.StoredKey))
			t.Log(string(cred.ServerKey))
		})
	}
}
