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
	"crypto/subtle"
	"io"

	"golang.org/x/crypto/argon2"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// plainParams represent parameters for PLAIN authentication using Argon2id.
//
// See https://www.rfc-editor.org/rfc/rfc9106.html#name-argon2-inputs-and-outputs.
type plainParams struct {
	t       uint32 // number of passes a.k.a. time
	p       uint8  // degree of parallelism a.k.a. threads
	m       uint32 // memory size in KiB
	tagLen  uint32 // output length a.k.a. key length
	saltLen int    // salt length
}

// fixedPlainParams represent fixed parameters for PLAIN authentication using Argon2id.
//
// We use values recommended by https://www.rfc-editor.org/rfc/rfc9106.html#section-4-6.2.
// Some other sources, such as https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
// or https://www.ietf.org/archive/id/draft-ietf-kitten-password-storage-04.html#name-argon2,
// recommend lower parameters.
// We also could automatically tweak parameters based on the available CPU and memory,
// but let's keep it simple for now.
var fixedPlainParams = plainParams{
	t:       3,
	p:       4,
	m:       64 * 1024,
	tagLen:  32,
	saltLen: 16,
}

// plainHashParams hashes the password using Argon2id with the given salt and parameters,
// and returns the document that should be stored and the hash.
func plainHashParams(password string, salt []byte, params *plainParams) (*types.Document, []byte, error) {
	if len(salt) != int(params.saltLen) {
		return nil, nil, lazyerrors.Errorf("unexpected salt length: %d", len(salt))
	}

	hash := argon2.IDKey([]byte(password), salt, params.t, params.m, params.p, params.tagLen)

	if len(hash) != int(params.tagLen) {
		return nil, nil, lazyerrors.Errorf("unexpected hash length: %d", len(hash))
	}

	doc, err := types.NewDocument(
		"algo", "argon2id",
		"t", int32(params.t),
		"p", int32(params.p),
		"m", int32(params.m),
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

// PlainHash hashes the password using Argon2id and returns the document that should be stored.
func PlainHash(password string) (*types.Document, error) {
	salt := make([]byte, fixedPlainParams.saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, _, err := plainHashParams(password, salt, &fixedPlainParams)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// plainVerifyParams verifies the password against the document returned by plainHashParams.
//
// Parameters are returned if they could be decoded from the document, even on error.
func plainVerifyParams(password string, doc *types.Document) (*plainParams, error) {
	v, _ := doc.Get("algo")
	algo, _ := v.(string)
	if algo != "argon2id" {
		return nil, lazyerrors.Errorf("invalid algo: %q", algo)
	}

	v, _ = doc.Get("t")
	t, _ := v.(int32)

	v, _ = doc.Get("p")
	p, _ := v.(int32)

	v, _ = doc.Get("m")
	m, _ := v.(int32)

	v, _ = doc.Get("hash")
	hash, _ := v.(types.Binary)

	v, _ = doc.Get("salt")
	salt, _ := v.(types.Binary)

	params := &plainParams{
		t:       uint32(t),
		p:       uint8(p),
		m:       uint32(m),
		tagLen:  uint32(len(hash.B)),
		saltLen: len(salt.B),
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
	if params != nil && *params != fixedPlainParams {
		return lazyerrors.Errorf("invalid params: %+v", params)
	}

	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
