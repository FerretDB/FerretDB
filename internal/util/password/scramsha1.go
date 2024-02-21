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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// SCRAMSHA1Hash computes SCRAM-SHA-1 credentials and returns the document that should be stored.
func SCRAMSHA1Hash(username string, password Password) (*types.Document, error) {
	salt := make([]byte, fixedScramSHA1Params.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := scramSHA1HashParams(username, password, salt, fixedScramSHA1Params)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// fixedScramSHA1Params represent fixed password parameters for SCRAM-SHA-1 authentication.
var fixedScramSHA1Params = &scramParams{
	iterationCount: 10_000,
	saltLen:        18,
}

// scramSHA1HashParams hashes the password with the given salt and parameters,
// and returns the document that should be stored using a variation of the SCRAM-SHA-1 algorithm
// used by MongoDB.
func scramSHA1HashParams(username string, password Password, salt []byte, params *scramParams) (*types.Document, error) {
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
