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
	"encoding/base64"
	"hash"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Computes the HMAC of the given data using the given key.
func computeHMAC(h func() hash.Hash, key, data []byte) []byte {
	mac := hmac.New(h, key)
	mac.Write(data)

	return mac.Sum(nil)
}

// Computes the SHA-1 hash of the given data.
func computeHash(h func() hash.Hash, b []byte) []byte {
	dh := h()
	dh.Write(b)

	return dh.Sum(nil)
}

// scramParams represent password parameters for SCRAM-SHA-1 authentication.
type scramParams struct {
	iterationCount int
	saltLen        int
}

// scramDoc creates a document with the stored key, iteration count, salt, and server key.
func scramDoc(h func() hash.Hash, saltedPassword, salt []byte, params *scramParams) (*types.Document, error) {
	clientKey := computeHMAC(h, saltedPassword, []byte("Client Key"))
	serverKey := computeHMAC(h, saltedPassword, []byte("Server Key"))
	storedKey := computeHash(h, clientKey)

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
