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

package sessions

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/internal/types"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestEndSessionsCommand(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/1554")

	s := setup.SetupWithOpts(t, nil)
	ctx := s.Ctx
	db := s.Collection.Database()

	for name, tc := range map[string]struct {
		command any
		err     *mongo.CommandError // nil if no error is expected
	}{
		"nilSessionID": {
			command: bson.A{bson.D{{"id", nil}}},
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: `BSON field 'endSessions.endSessionsFromClient.id' is missing but a required field`,
			},
		},
		"invalidSessionID": {
			command: bson.A{bson.D{{"id", "invalid"}}},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: `BSON field 'endSessions.endSessionsFromClient.id' is the wrong type 'string', expected type 'binData'`,
			},
		},
		"nonexistentSessionID": {
			command: bson.A{bson.D{{
				"id", primitive.Binary{Subtype: byte(types.BinaryUUID), Data: []byte{
					0x01, 0x01, 0x01, 0x01,
					0x01, 0x01, 0x01, 0x01,
					0x01, 0x01, 0x01, 0x01,
					0x01, 0x01, 0x01, 0x01,
				}},
			}}},
		},
		"nilCommand": {
			command: nil,
			err: &mongo.CommandError{
				Code:    10065,
				Name:    "Location10065",
				Message: `invalid parameter: expected an object (endSessions)`,
			},
		},
		"emptyCommand": {
			command: bson.A{},
		},
	} {
		name, tc := name, tc

		tt.Run(name, func(t *testing.T) {
			t.Parallel()

			var res bson.D
			endSessionsCommand := bson.D{{"endSessions", tc.command}}
			err := db.RunCommand(ctx, endSessionsCommand).Decode(&res)

			if tc.err != nil {
				integration.AssertEqualAltCommandError(t, *tc.err, "", err)
				return
			}

			require.NoError(t, err)

			doc := integration.ConvertDocument(t, res)
			assert.Equal(t, float64(1), must.NotFail(doc.Get("ok")))
		})
	}
}
