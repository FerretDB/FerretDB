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
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"

	"golang.org/x/crypto/pbkdf2"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// SCRAMSHA1VariationHash computes a variation of SCRAM-SHA-1 and returns
// a document containing stored key, iteration count, salt, and server key.
//
// It does not conform to the SCRAM-SHA-1 standard due to the custom preparation
// of the password.
func SCRAMSHA1VariationHash(username string, password Password) (*bson.Document, error) {
	salt := make([]byte, fixedScramSHA1Params.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := scramSHA1VariationHashParams(username, password, salt, fixedScramSHA1Params)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// fixedScramSHA1Params represent fixed password parameters for SCRAM-SHA-1 authentication.
var fixedScramSHA1Params = &scramParams{
	iterationCount: 10_000,
	saltLen:        16,
}

// scramSHA1VariationHashParams hashes the password using custom preparation with the given salt and parameters,
// and returns the document that should be stored using a variation of the SCRAM-SHA-1 algorithm
// used by MongoDB.
func scramSHA1VariationHashParams(username string, password Password, salt []byte, params *scramParams) (*bson.Document, error) {
	if len(salt) != int(params.saltLen) {
		return nil, lazyerrors.Errorf("unexpected salt length: %d", len(salt))
	}

	md5sum := md5.New()
	if _, err := md5sum.Write([]byte(username + ":mongo:" + password.Password())); err != nil {
		return nil, lazyerrors.Error(err)
	}

	src := md5sum.Sum(nil)
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)

	saltedPassword := pbkdf2.Key(dst, salt, params.iterationCount, sha1.Size, sha1.New)

	return scramDoc(sha1.New, saltedPassword, salt, params)
}
