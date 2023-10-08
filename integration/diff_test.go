// Copyright 2023 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package integration

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestInsertDuplicateKeys(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName:   "testduplicatekeys",
		CollectionName: "insert-duplicate-keys",
	})
	ctx, collection := s.Ctx, s.Collection

	tests := []struct {
		name        string
		expectedErr mongo.WriteException
		doc         interface{}
	}{
		{
			name: "simple duplicate key",
			doc:  bson.D{{"_id", "duplicate_keys"}, {"foo", "bar"}, {"foo", "baz"}},
			expectedErr: mongo.WriteException{WriteErrors: []mongo.WriteError{{
				Index:   0,
				Code:    2,
				Message: `invalid key: "foo" (duplicate keys are not allowed)`,
			}}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := collection.InsertOne(ctx, tc.doc)
			assert.Equal(t, tc.expectedErr, unsetRaw(t, err))
		})
	}
}

func TestNestedArrays(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName:   "testnestedarrays",
		CollectionName: "nested-arrays",
	})
	ctx, collection := s.Ctx, s.Collection

	tests := []struct {
		name        string
		expectedErr error
		testFn      func(t *testing.T) error
	}{
		{
			name: "insert with nested arrays",
			expectedErr: mongo.WriteException{WriteErrors: []mongo.WriteError{{
				Index:   0,
				Code:    2,
				Message: `invalid value: { "foo": [ [ "bar" ] ] } (nested arrays are not supported)`,
			}}},
			testFn: func(t *testing.T) error {
				_, err := collection.InsertOne(ctx, bson.D{{"foo", bson.A{bson.A{"bar"}}}})
				return err
			},
		},
		{
			name: "update with nested arrays",
			expectedErr: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `invalid value: { "foo1": [ [ "bar1" ] ] } (nested arrays are not supported)`,
			},
			testFn: func(t *testing.T) error {
				// initiate a collection with a valid document, so we have something to update
				_, err := collection.InsertOne(ctx, bson.D{{"_id", "valid"}, {"v", int32(42)}})
				require.NoError(t, err)

				_, err = collection.UpdateOne(ctx, bson.D{}, bson.D{{"$set", bson.D{{"foo1", bson.A{bson.A{"bar1"}}}}}})
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.testFn(t)
			assert.Equal(t, tc.expectedErr, unsetRaw(t, err))
		})
	}
}

func TestDatabaseCollectionName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		dbName         string
		collectionName string
		expectedErr    error
	}{
		{
			name:           "DB name with reserved prefix",
			dbName:         "_ferretdb_xxx",
			collectionName: "test",
			expectedErr: mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: `Invalid namespace specified '_ferretdb_xxx.test'`,
			},
		},
		{
			name:           "DB name with non-latin characters",
			dbName:         "データベース",
			collectionName: "test",
			expectedErr: mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: `Invalid namespace specified 'データベース.test'`,
			},
		},
		{
			name:           "Collection name with reserved prefix",
			dbName:         "testcollectionnamereservedprefix",
			collectionName: "_ferretdb_xxx",
			expectedErr: mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid collection name: %s`, "_ferretdb_xxx"),
			},
		},
		{
			name:           "Collection name with non-latin characters",
			dbName:         "testcollectionnamenonlatichars",
			collectionName: string([]byte{0xff, 0xfe, 0xfd}),
			expectedErr: mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid collection name: %s`, string([]byte{0xff, 0xfe, 0xfd})),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			setup.SetupWithOpts(t, &setup.SetupOpts{
				DatabaseName:   tc.dbName,
				CollectionName: tc.collectionName,
			})
			// TODO: run test and see how we can verify
		})
	}
}

func TestNullStrings(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)
	testCases := []struct {
		name               string
		doc                interface{}
		expectedErrPattern string
	}{
		{
			name:               "Null Strings value",
			doc:                bson.D{{"_id", "document"}, {"a", string([]byte{0})}},
			expectedErrPattern: "^.* unsupported Unicode escape sequence .*$",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := collection.InsertOne(ctx, tc.doc)
			require.Error(t, err)
			assert.Regexp(t, tc.expectedErrPattern, err.Error())
		})
	}
}

func TestNegativeZero(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName:   "testnegativezero",
		CollectionName: "negative-zero",
	})
	ctx, collection := s.Ctx, s.Collection

	testCases := []struct {
		name      string
		insertDoc interface{}
		updateDoc interface{}
		filter    interface{}
	}{
		{
			name:      "Insert negative zero",
			insertDoc: bson.D{{"_id", "1"}, {"foo", math.Copysign(0.0, -1)}},
			filter:    bson.D{{"_id", "1"}},
		},
		{
			name:      "Multiply zero by negative",
			insertDoc: bson.D{{"_id", "zero"}, {"v", int32(0)}},
			updateDoc: bson.D{{"$mul", bson.D{{"v", float64(-1)}}}},
			filter:    bson.D{{"_id", "zero"}},
		},
		{
			name:      "Multiple negative by zero",
			insertDoc: bson.D{{"_id", "negative"}, {"v", int64(-1)}},
			updateDoc: bson.D{{"$mul", bson.D{{"v", float64(0)}}}},
			filter:    bson.D{{"_id", "negative"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := collection.InsertOne(ctx, tc.insertDoc)
			require.NoError(t, err)

			if tc.updateDoc != nil {
				_, err = collection.UpdateOne(ctx, tc.filter, tc.updateDoc)
				require.NoError(t, err)
			}

			var res bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&res)
			require.NoError(t, err)

			var actual float64
			for _, e := range res {
				if e.Key == "foo" {
					var ok bool
					actual, ok = e.Value.(float64)
					require.True(t, ok)
				}
			}
			require.Equal(t, 0.0, actual)
			require.Equal(t, math.Signbit(math.Copysign(0.0, +1)), math.Signbit(actual))
		})
	}
}

func TestUpdateProduceInfinity(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName:   "testupdateproduceinfiity",
		CollectionName: "produce-infinity",
	})
	ctx, collection := s.Ctx, s.Collection

	testCases := []struct {
		name     string
		filter   bson.D
		insert   bson.D
		update   bson.D
		expected mongo.CommandError
	}{
		{
			name:   "MulInf",
			filter: bson.D{{"_id", "number"}},
			insert: bson.D{{"_id", "number"}, {"v", int32(42)}},
			update: bson.D{{"$mul", bson.D{{"v", math.MaxFloat64}}}},
			expected: mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `update produces invalid value: { "v": +Inf }` +
					` (update operations that produce infinity values are not allowed)`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := collection.InsertOne(ctx, tc.insert)
			require.NoError(t, err)

			_, err = collection.UpdateOne(ctx, tc.filter, tc.update)

			assertEqualError(t, tc.expected, err)
		})
	}
}

func TestDocumentValidation(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName:   "testdocumentvalidation",
		CollectionName: "doc-validation",
	})
	ctx, collection := s.Ctx, s.Collection
	t.Run("Insert", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			name     string
			doc      bson.D
			expected error
		}{
			{
				name: "DollarSign",
				doc:  bson.D{{"$foo", "bar"}},
				expected: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Index:   0,
					Code:    2,
					Message: `invalid key: "$foo" (key must not start with '$' sign)`,
				}}},
			},
			{
				name: "DotSign",
				doc:  bson.D{{"foo.bar", "baz"}},
				expected: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Index:   0,
					Code:    2,
					Message: `invalid key: "foo.bar" (key must not contain '.' sign)`,
				}}},
			},
			{
				name: "Infinity",
				doc:  bson.D{{"_id", "1"}, {"foo", math.Inf(1)}},
				expected: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Code:    2,
					Message: `invalid value: { "foo": +Inf } (infinity values are not allowed)`,
				}}},
			},
			{
				name: "NegativeInfinity",
				doc:  bson.D{{"_id", "1"}, {"foo", math.Inf(-1)}},
				expected: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Code:    2,
					Message: `invalid value: { "foo": -Inf } (infinity values are not allowed)`,
				}}},
			},
			{
				name: "NaN",
				doc:  bson.D{{"_id", "1"}, {"foo", math.NaN()}},
				expected: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { insert: "doc-validation", ordered: true, ` +
						`$db: "testdocumentvalidation", documents: [ { _id: "1", foo: nan.0 } ] } with: NaN is not supported`,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				_, err := collection.InsertOne(ctx, tc.doc)
				assert.Equal(t, tc.expected, unsetRaw(t, err))
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name     string
			doc      bson.D
			expected mongo.CommandError
		}{
			{
				name: "",
				doc:  bson.D{{"$set", bson.D{{"foo", bson.D{{"bar.baz", "qaz"}}}}}},
				expected: mongo.CommandError{
					Code:    2,
					Name:    "BadValue",
					Message: `invalid key: "bar.baz" (key must not contain '.' sign)`,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// initiate a collection with a valid document, so we have something to update
				_, err := collection.InsertOne(ctx, bson.D{{"_id", "valid"}, {"v", int32(42)}})
				require.NoError(t, err)

				_, err = collection.UpdateOne(ctx, bson.D{}, tc.doc)
				assertEqualError(t, tc.expected, err)
			})
		}
	})

	t.Run("FindAndModify", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name     string
			command  bson.D
			expected mongo.CommandError
		}{
			{
				name: "DollarPrefixFieldName",
				command: bson.D{
					{"query", bson.D{{"_id", bson.D{{"k", bson.D{{"$invalid", "v"}}}}}}},
					{"upsert", true},
					{"update", bson.D{{"v", "replaced"}}},
				},
				expected: mongo.CommandError{
					Code: 52,
					Name: "DollarPrefixedFieldName",
					Message: `Plan executor error during findAndModify :: caused by :: ` +
						`_id fields may not contain '$'-prefixed fields: $invalid is not valid for storage.`,
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				command := bson.D{{"findAndModify", collection.Name()}}
				command = append(command, tc.command...)

				err := collection.Database().RunCommand(ctx, command)

				assertEqualError(t, tc.expected, err.Err())
			})
		}
	})
}

func TestDBStatsScale(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "testerrormgs",
	})
	ctx, db := s.Ctx, s.Collection.Database()

	testCases := []struct {
		name                string
		scale               any
		expectedMongoDBErr  string
		expectedFerretDBErr mongo.CommandError
	}{
		{
			name:               "Zero",
			scale:              int32(0),
			expectedMongoDBErr: "scale has to be > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "Location51024",
				Code:    51024,
				Message: "BSON field 'scale' value must be >= 1, actual value '0'",
			},
		},
		{
			name:               "Negative",
			scale:              int32(-100),
			expectedMongoDBErr: "scale has to be > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "Location51024",
				Code:    51024,
				Message: "BSON field 'scale' value must be >= 1, actual value '-100'",
			},
		},
		{
			name:               "MinFloat",
			scale:              -math.MaxFloat64,
			expectedMongoDBErr: "scale has to be > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "Location51024",
				Code:    51024,
				Message: "BSON field 'scale' value must be >= 1, actual value '-9223372036854775808'",
			},
		},
		{
			name:               "String",
			scale:              "1",
			expectedMongoDBErr: "scale has to be a number > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "TypeMismatch",
				Code:    14,
				Message: "BSON field 'dbStats.scale' is the wrong type 'string', expected types '[long, int, decimal, double]'",
			},
		},
		{
			name:               "Object",
			scale:              bson.D{{"a", 1}},
			expectedMongoDBErr: "scale has to be a number > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "TypeMismatch",
				Code:    14,
				Message: "BSON field 'dbStats.scale' is the wrong type 'object', expected types '[long, int, decimal, double]'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := db.RunCommand(ctx, bson.D{{"dbStats", int32(1)}, {"scale", tc.scale}}).Err()
			require.Error(t, err)
			assertEqualError(t, tc.expectedFerretDBErr, err)
		})
	}
}

func TestErrorMessages(t *testing.T) {
	t.Parallel()
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "testerrormgs",
	})
	ctx, db := s.Ctx, s.Collection.Database()

	testCases := []struct {
		name   string
		testFn func(t *testing.T)
	}{
		{
			name: "getParameter",
			testFn: func(t *testing.T) {
				var doc bson.D
				err := db.RunCommand(ctx, bson.D{{"getParameter", bson.D{{"allParameters", "1"}}}}).Decode(&doc)

				var actual mongo.CommandError
				require.ErrorAs(t, err, &actual)
				assert.Equal(t, int32(14), actual.Code)
				assert.Equal(t, "TypeMismatch", actual.Name)
				expectedErr := mongo.CommandError{
					Code:    14,
					Name:    "TypeMismatch",
					Message: "BSON field 'allParameters' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
				}
				assertEqualError(t, expectedErr, actual)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.testFn(t)
		})
	}
}
