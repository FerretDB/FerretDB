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
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// NewSCRAMSHA256 computes SCRAM-SHA-256 credentials and returns the document that should be stored.
func NewSCRAMSHA256(username, password string) (*types.Document, error) {
	client, err := scram.SHA256.NewClient(username, password, "")
	if err != nil {
		return nil, err
	}

	salt := make([]byte, 45)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	kf := scram.KeyFactors{
		Salt:  string(salt),
		Iters: 15000,
	}

	cred := client.GetStoredCredentials(kf)

	doc := must.NotFail(types.NewDocument())
	doc.Set("storedKey", base64.StdEncoding.EncodeToString(cred.StoredKey))
	doc.Set("iterationCount", int32(kf.Iters))
	doc.Set("salt", base64.StdEncoding.EncodeToString(salt))
	doc.Set("serverKey", base64.StdEncoding.EncodeToString(cred.ServerKey))

	return doc, nil
}
