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
	"testing"

	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

func TestCheckNoTransaction(t *testing.T) {
	t.Parallel()

	lsid := wirebson.MustDocument("id", wirebson.Binary{B: make([]byte, 16), Subtype: wirebson.BinaryUUID})

	for name, tc := range map[string]struct {
		doc      *wirebson.Document
		argument string // empty if the command is allowed
	}{
		"NoSession": {
			doc: wirebson.MustDocument("insert", "c", "$db", "test"),
		},
		"Session": {
			doc: wirebson.MustDocument("insert", "c", "lsid", lsid, "$db", "test"),
		},
		"RetryableWrite": {
			doc: wirebson.MustDocument("insert", "c", "lsid", lsid, "txnNumber", int64(1), "$db", "test"),
		},
		"StartTransaction": {
			doc: wirebson.MustDocument(
				"insert", "c",
				"lsid", lsid,
				"txnNumber", int64(1),
				"startTransaction", true,
				"autocommit", false,
				"$db", "test",
			),
			argument: "startTransaction",
		},
		"InTransaction": {
			doc: wirebson.MustDocument(
				"update", "c",
				"lsid", lsid,
				"txnNumber", int64(1),
				"autocommit", false,
				"$db", "test",
			),
			argument: "autocommit",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := checkNoTransaction(tc.doc)

			if tc.argument == "" {
				assert.Nil(t, err)
				return
			}

			require.NotNil(t, err)
			assert.Equal(t, int32(mongoerrors.ErrNotImplemented), err.Code)
			assert.Equal(t, tc.argument, err.Argument)
		})
	}
}
