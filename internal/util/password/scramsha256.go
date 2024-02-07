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

	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// scramSHA256Params represent password parameters for SCRAM-SHA-256 authentication.
type scramSHA256Params struct {
	iterationCount int
	saltLen        int
}

// fixedScramSHA256Params represent fixed password parameters for SCRAM-SHA-256 authentication.
var fixedScramSHA256Params = &scramSHA256Params{
	iterationCount: 15_000,
	saltLen:        30,
}

// scramSHA256HashParams hashes the password with the given salt and parameters,
// and returns the document that should be stored.
func scramSHA256HashParams(password string, salt []byte, params *scramSHA256Params) (*types.Document, error) {
	if len(salt) != int(params.saltLen) {
		return nil, lazyerrors.Errorf("unexpected salt length: %d", len(salt))
	}

	// username is omitted because it is not used in the hash computation.
	client, err := scram.SHA256.NewClient("", password, "")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	kf := scram.KeyFactors{
		Salt:  string(salt),
		Iters: params.iterationCount,
	}

	cred := client.GetStoredCredentials(kf)

	doc, err := types.NewDocument(
		"storedKey", base64.StdEncoding.EncodeToString(cred.StoredKey),
		"iterationCount", int32(params.iterationCount),
		"salt", base64.StdEncoding.EncodeToString(salt),
		"serverKey", base64.StdEncoding.EncodeToString(cred.ServerKey),
	)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// SCRAMSHA256Hash computes SCRAM-SHA-256 credentials and returns the document that should be stored.
func SCRAMSHA256Hash(password string) (*types.Document, error) {
	salt := make([]byte, fixedScramSHA256Params.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := scramSHA256HashParams(password, salt, fixedScramSHA256Params)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}
