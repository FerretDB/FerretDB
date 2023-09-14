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

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

func TestInsertCommandErrors(tt *testing.T) {
	tt.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		toInsert []any // required, slice of bson.D to insert
		ordered  any   // required, sets it to `ordered`

		cerr           *mongo.CommandError // optional, expected command error from MongoDB
		werr           *mongo.WriteError   // optional, expected write error from MongoDB
		altMessage     string              // optional, alternative error message for FerretDB, ignored if empty
		skip           string              // optional, skip test with a specified reason
		failsForSQLite string              // optional, if set, the case is expected to fail for SQLite due to given issue
	}{
		"InsertOrderedInvalid": {
			toInsert: []any{
				bson.D{{"_id", "foo"}},
			},
			ordered: "foo",
			cerr: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'insert.ordered' is the wrong type 'string', expected type 'bool'",
			},
			altMessage: "BSON field 'ordered' is the wrong type 'string', expected type 'bool'",
		},
		"InsertDuplicateKey": {
			toInsert: []any{
				bson.D{{"_id", "double"}},
			},
			ordered: false,
			werr: &mongo.WriteError{
				Code: 11000,
				Message: `E11000 duplicate key error collection: ` +
					`TestInsertCommandErrors-InsertDuplicateKey.TestInsertCommandErrors-InsertDuplicateKey index: _id_ dup key: { _id: "double" }`,
			},
			altMessage: "E11000 duplicate key error collection: TestInsertCommandErrors-InsertDuplicateKey.TestInsertCommandErrors-InsertDuplicateKey",
		},
		"InsertDuplicateKeyOrdered": {
			toInsert: []any{
				bson.D{{"foo", "bar"}},
				bson.D{{"_id", "double"}},
			},
			ordered: true,
			werr: &mongo.WriteError{
				Code:  11000,
				Index: 1,
				Message: `E11000 duplicate key error collection: ` +
					`TestInsertCommandErrors-InsertDuplicateKeyOrdered.TestInsertCommandErrors-InsertDuplicateKeyOrdered index: _id_ dup key: { _id: "double" }`,
			},
			altMessage: `E11000 duplicate key error collection: TestInsertCommandErrors-InsertDuplicateKeyOrdered.TestInsertCommandErrors-InsertDuplicateKeyOrdered`,
		},
		"InsertArray": {
			toInsert: []any{
				bson.D{{"a", int32(1)}},
				bson.A{},
			},
			ordered: true,
			cerr: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'insert.documents.1' is the wrong type 'array', expected type 'object'",
			},
		},
	} {
		name, tc := name, tc
		tt.Run(name, func(tt *testing.T) {
			if tc.skip != "" {
				tt.Skip(tc.skip)
			}

			tt.Parallel()

			var t testtb.TB = tt
			if tc.failsForSQLite != "" {
				t = setup.FailsForSQLite(tt, tc.failsForSQLite)
			}

			require.NotNil(t, tc.toInsert, "toInsert must not be nil")
			require.NotNil(t, tc.ordered, "ordered must not be nil")

			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			var res bson.D
			err := collection.Database().RunCommand(ctx, bson.D{
				{"insert", collection.Name()},
				{"documents", tc.toInsert},
				{"ordered", tc.ordered},
			}).Decode(&res)

			assert.Nil(t, res)

			if tc.cerr != nil {
				AssertEqualAltCommandError(t, *tc.cerr, tc.altMessage, err)
				return
			}

			if tc.werr != nil {
				AssertEqualAltWriteError(t, *tc.werr, tc.altMessage, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestInsertIDDifferentTypes(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	_, err := collection.InsertOne(ctx, bson.D{
		{"_id", int64(1)},
		{"v", "foo2"},
	})
	require.NoError(t, err)

	_, err = collection.InsertOne(ctx, bson.D{
		{"_id", int32(1)},
		{"v", "foo1"},
	})

	AssertEqualAltWriteError(t, mongo.WriteError{
		Message: "E11000 duplicate key error collection: TestInsertIDDifferentTypes.TestInsertIDDifferentTypes index: _id_ dup key: { _id: 1 }",
		Code:    11000,
	},
		"E11000 duplicate key error collection: TestInsertIDDifferentTypes.TestInsertIDDifferentTypes",
		err)

	_, err = collection.InsertOne(ctx, bson.D{
		{"_id", float32(1)},
		{"v", "foo3"},
	})

	AssertEqualAltWriteError(t, mongo.WriteError{
		Message: "E11000 duplicate key error collection: TestInsertIDDifferentTypes.TestInsertIDDifferentTypes index: _id_ dup key: { _id: 1.0 }",
		Code:    11000,
	},
		"E11000 duplicate key error collection: TestInsertIDDifferentTypes.TestInsertIDDifferentTypes",
		err)
}
