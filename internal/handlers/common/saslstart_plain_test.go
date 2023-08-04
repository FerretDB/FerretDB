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

package common

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// TestSaslStartPlain tests saslStartPlain function.
// Integration tests are not possible because the driver
// used in integration tests doesn't support all possible scenarios.
// To ensure compatibility, the functionality must be tested by external drivers
// of various programming languages (see the dance repo).
func TestSaslStartPlain(t *testing.T) {
	validPayload := []byte("authzid\x00admin\x00pass")

	for name, tc := range map[string]struct { //nolint:vet // for readability
		doc *types.Document

		// expected results
		username string
		password string
		err      error
	}{
		"emptyPayload": {
			doc: types.MakeDocument(0),
			err: commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				`required parameter "payload" is missing`,
				"payload",
			),
		},
		"wrongTypePayload": {
			doc: must.NotFail(types.NewDocument("payload", 42)),
			err: commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				`required parameter "payload" has type int (expected types.Binary)`,
				"payload",
			),
		},
		"stringPayloadInvalid": {
			doc: must.NotFail(types.NewDocument("payload", "ABC")),
			err: commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				"Invalid payload: illegal base64 data at input byte 0",
				"payload",
			),
		},
		"binaryPayloadInvalid": {
			doc: must.NotFail(types.NewDocument("payload", types.Binary{B: []byte("ABC")})),
			err: commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				"Invalid payload: expected 3 parts, got 1",
				"payload",
			),
		},
		"stringPayload": {
			doc:      must.NotFail(types.NewDocument("payload", base64.StdEncoding.EncodeToString(validPayload))),
			username: "admin",
			password: "pass",
			err:      nil,
		},
		"binaryPayload": {
			doc:      must.NotFail(types.NewDocument("payload", types.Binary{B: validPayload})),
			username: "admin",
			password: "pass",
			err:      nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			username, password, err := saslStartPlain(tc.doc)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.username, username)
			assert.Equal(t, tc.password, password)
		})
	}
}
