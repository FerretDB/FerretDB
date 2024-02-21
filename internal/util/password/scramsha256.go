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
	"fmt"

	"github.com/xdg-go/stringprep"
	"golang.org/x/crypto/pbkdf2"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// SCRAMSHA256Hash computes SCRAM-SHA-256 credentials and returns the document that should be stored.
func SCRAMSHA256Hash(password string) (*types.Document, error) {
	salt := make([]byte, fixedScramSHA256Params.saltLen-4) // minus 4 to base64 length to 40 bytes.
	if _, err := rand.Read(salt); err != nil {
		return nil, lazyerrors.Error(err)
	}

	suffix := make([]byte, 4)
	suffix[0] = 0
	suffix[1] = 0
	suffix[2] = 0
	suffix[3] = 1

	salt = append(salt, suffix...)

	doc, err := scramSHA256HashParams(password, salt, fixedScramSHA256Params)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// fixedScramSHA256Params represent fixed password parameters for SCRAM-SHA-256 authentication.
var fixedScramSHA256Params = &scramParams{
	iterationCount: 15_000,
	saltLen:        28, // 32 byte SHA-256 block size minus 4 bytes for suffix.
}

// scramSHA256HashParams hashes the password with the given salt and parameters,
// and returns the document that should be stored.
//
// https://datatracker.ietf.org/doc/html/rfc5802
func scramSHA256HashParams(password string, salt []byte, params *scramParams) (*types.Document, error) {
	if len(salt) != int(params.saltLen) {
		return nil, lazyerrors.Errorf("unexpected salt length: %d", len(salt))
	}

	prepPassword, err := stringprep.SASLprep.Prepare(password)
	if err != nil {
		return nil, fmt.Errorf("cannot SASLprepare password '%s': %v", password, err)
	}

	saltedPassword := pbkdf2.Key([]byte(prepPassword), salt, params.iterationCount, sha256.Size, sha256.New)

	fmt.Println(len(saltedPassword))

	return scramDoc(sha256.New, saltedPassword, salt, params)
}
