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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/xdg-go/stringprep"
	"golang.org/x/crypto/pbkdf2"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

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
//
// https://datatracker.ietf.org/doc/html/rfc5802
func scramSHA256HashParams(password string, salt []byte, params *scramSHA256Params) (*types.Document, error) {
	if len(salt) != int(params.saltLen) {
		return nil, lazyerrors.Errorf("unexpected salt length: %d", len(salt))
	}

	prepPassword, err := stringprep.SASLprep.Prepare(password)
	if err != nil {
		return nil, fmt.Errorf("cannot SASLprepare password '%s': %v", password, err)
	}

	saltedPassword := pbkdf2.Key([]byte(prepPassword), salt, params.iterationCount, sha256.Size, sha256.New)

	// Hashing the strings "Client Key" for creating the client key and
	// "Server Key" for creating the server key, per Section 3 of the RFC 5802
	// https://datatracker.ietf.org/doc/html/rfc5802#section-3
	clientKey := computeHMAC(saltedPassword, []byte("Client Key"))
	serverKey := computeHMAC(saltedPassword, []byte("Server Key"))

	storedKey := computeHash(clientKey)

	doc, err := types.NewDocument(
		"storedKey", base64.StdEncoding.EncodeToString(storedKey),
		"iterationCount", int32(params.iterationCount),
		"salt", base64.StdEncoding.EncodeToString(salt),
		"serverKey", base64.StdEncoding.EncodeToString(serverKey),
	)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// Computes the HMAC of the given data using the given key.
func computeHMAC(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)

	return mac.Sum(nil)
}

// Computes the SHA-256 hash of the given data.
func computeHash(b []byte) []byte {
	h := sha256.New()
	h.Write(b)

	return h.Sum(nil)
}
