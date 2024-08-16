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

package handler

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
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
		db       string
		err      error
	}{
		"emptyPayload": {
			doc: types.MakeDocument(0),
			err: handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrBadValue,
				`required parameter "payload" is missing`,
				"payload",
			),
		},
		"wrongTypePayload": {
			doc: must.NotFail(types.NewDocument("payload", int32(42))),
			err: handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrBadValue,
				`required parameter "payload" has type int32 (expected types.Binary)`,
				"payload",
			),
		},
		"stringPayloadInvalid": {
			doc: must.NotFail(types.NewDocument("payload", "ABC")),
			err: handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrBadValue,
				"Invalid payload: illegal base64 data at input byte 0",
				"payload",
			),
		},
		"binaryPayloadInvalid": {
			doc: must.NotFail(types.NewDocument("payload", types.Binary{B: []byte("ABC")})),
			err: handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrTypeMismatch,
				"Invalid payload: expected 3 fields, got 1",
				"payload",
			),
		},
		"stringPayload": {
			doc:      must.NotFail(types.NewDocument("payload", base64.StdEncoding.EncodeToString(validPayload))),
			username: "admin",
			password: "pass",
			db:       "db",
		},
		"binaryPayload": {
			doc:      must.NotFail(types.NewDocument("payload", types.Binary{B: validPayload})),
			username: "admin",
			password: "pass",
			db:       "db",
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
			err := saslStartPlain(ctx, "db", tc.doc)
			assert.Equal(t, tc.err, err)

			username, password, _, db := conninfo.Get(ctx).Auth()
			assert.Equal(t, tc.username, username)
			assert.Equal(t, tc.password, password)
			assert.Equal(t, tc.db, db)
		})
	}
}
