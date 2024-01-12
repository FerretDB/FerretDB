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
	"crypto/sha256"
	"crypto/subtle"
	"io"

	"golang.org/x/crypto/pbkdf2"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// The mechanism and the (only) algorithm used for PLAIN authentication.
const (
	plainMechanism = "PLAIN"
	plainAlgo      = "PBKDF2-HMAC-SHA256"
)

// plainParams represent password parameters for PLAIN authentication.
type plainParams struct {
	iterationCount int
	saltLen        int
	hashLen        int
}

// fixedPlainParams represent fixed password parameters for PLAIN authentication using PBKDF2-HMAC-SHA256.
//
// See https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#pbkdf2.
var fixedPlainParams = &plainParams{
	iterationCount: 600_000,
	saltLen:        16,
	hashLen:        32,
}

// plainHashParams hashes the password with the given salt and parameters,
// and returns the document that should be stored and the hash.
func plainHashParams(password string, salt []byte, params *plainParams) (*types.Document, []byte, error) {
	if len(salt) != int(params.saltLen) {
		return nil, nil, lazyerrors.Errorf("unexpected salt length: %d", len(salt))
	}

	hash := pbkdf2.Key([]byte(password), salt, params.iterationCount, params.hashLen, sha256.New)

	if len(hash) != int(params.hashLen) {
		return nil, nil, lazyerrors.Errorf("unexpected hash length: %d", len(hash))
	}

	doc, err := types.NewDocument(
		"mechanism", plainMechanism,
		"algo", plainAlgo,
		"iterationCount", int64(params.iterationCount),
		"hash", types.Binary{
			B: hash,
		},
		"salt", types.Binary{
			B: salt,
		},
	)
	if err != nil {
		return nil, nil, lazyerrors.Error(err)
	}

	return doc, hash, nil
}

// PlainHash hashes the password and returns the document that should be stored.
func PlainHash(password string) (*types.Document, error) {
	salt := make([]byte, fixedPlainParams.saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, _, err := plainHashParams(password, salt, fixedPlainParams)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// plainVerifyParams verifies the password against the document returned by plainHashParams.
//
// Parameters are returned if they could be decoded from the document,
// even if password is invalid or on any other error.
func plainVerifyParams(password string, doc *types.Document) (*plainParams, error) {
	v, _ := doc.Get("mechanism")
	mechanism, _ := v.(string)

	if mechanism != plainMechanism {
		return nil, lazyerrors.Errorf("invalid mechanism: %q", mechanism)
	}

	v, _ = doc.Get("algo")
	algo, _ := v.(string)

	if algo != plainAlgo {
		return nil, lazyerrors.Errorf("invalid algorithm: %q", algo)
	}

	v, _ = doc.Get("iterationCount")
	iterationCount, _ := v.(int64)

	v, _ = doc.Get("hash")
	hash, _ := v.(types.Binary)

	v, _ = doc.Get("salt")
	salt, _ := v.(types.Binary)

	params := &plainParams{
		iterationCount: int(iterationCount),
		saltLen:        len(salt.B),
		hashLen:        len(hash.B),
	}

	_, expectedHash, err := plainHashParams(password, salt.B, params)
	if err != nil {
		return params, lazyerrors.Error(err)
	}

	if subtle.ConstantTimeCompare(expectedHash, hash.B) != 1 {
		return params, lazyerrors.New("invalid password")
	}

	return params, nil
}

// PlainVerify verifies the password against the document returned by PlainHash.
//
// The returned error is safe for logging, but should not be exposed to the client.
func PlainVerify(password string, doc *types.Document) error {
	params, err := plainVerifyParams(password, doc)
	if params != nil && *params != *fixedPlainParams {
		return lazyerrors.Errorf("invalid params: %+v", params)
	}

	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
