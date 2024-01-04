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

package util

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/xdg-go/scram"
)

const IterationCount = 15000

type ScramConversation struct {
	Salt      string
	StoredKey []byte
	ServerKey []byte
	Conv      *scram.ServerConversation
	Mechanism string
}

func GenerateNonce() string {
	b := make([]byte, 24)
	rand.Read(b)
	nonce := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(nonce, b)
	return string(nonce)
}
