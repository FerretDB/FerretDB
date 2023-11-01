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
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestDiffInsertDuplicateKeys(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	doc := bson.D{{"_id", "duplicate_keys"}, {"foo", "bar"}, {"foo", "baz"}}
	_, err := collection.InsertOne(ctx, doc)

	if setup.IsMongoDB(t) {
		require.NoError(t, err)
		return
	}

	expected := mongo.WriteError{
		Index:   0,
		Code:    2,
		Message: `invalid key: "foo" (duplicate keys are not allowed)`,
	}
	AssertEqualWriteError(t, expected, err)
}

func TestDiffInsertObjectIDHexString(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	hex := "000102030405060708091011"

	objID, err := primitive.ObjectIDFromHex(hex)
	require.NoError(t, err)

	_, err = collection.InsertOne(ctx, bson.D{
		{"_id", objID},
	})
	require.NoError(t, err)

	_, err = collection.InsertOne(ctx, bson.D{
		{"_id", hex},
	})

	if setup.IsMongoDB(t) {
		require.NoError(t, err)
		return
	}

	expected := mongo.WriteError{
		Index:   0,
		Code:    11000,
		Message: `E11000 duplicate key error collection: TestDiffInsertObjectIDHexString.TestDiffInsertObjectIDHexString`,
	}
	AssertEqualWriteError(t, expected, err)
}

func TestDiffNestedArrays(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()

		_, err := collection.InsertOne(ctx, bson.D{{"foo", bson.A{bson.A{"bar"}}}})

		if setup.IsMongoDB(t) {
			require.NoError(t, err)
			return
		}

		expected := mongo.WriteError{
			Index:   0,
			Code:    2,
			Message: `invalid value: { "foo": [ [ "bar" ] ] } (nested arrays are not supported)`,
		}
		AssertEqualWriteError(t, expected, err)
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		_, err := collection.UpdateOne(ctx, bson.D{}, bson.D{{"$set", bson.D{{"foo", bson.A{bson.A{"bar"}}}}}})

		if setup.IsMongoDB(t) {
			require.NoError(t, err)
			return
		}

		expected := mongo.WriteError{
			Code:    2,
			Message: `invalid value: { "foo": [ [ "bar" ] ] } (nested arrays are not supported)`,
		}
		AssertEqualWriteError(t, expected, err)
	})
}

func TestDiffDatabaseName(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, dbName := range map[string]string{
		"ReservedPrefix": "_ferretdb_xxx",
		"NonLatin":       "データベース",
	} {
		name, dbName := name, dbName
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cName := testutil.CollectionName(t)

			err := collection.Database().Client().Database(dbName).CreateCollection(ctx, cName)

			if setup.IsMongoDB(t) {
				require.NoError(t, err)

				err = collection.Database().Client().Database(dbName).Drop(ctx)
				require.NoError(t, err)

				return
			}

			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid namespace specified '%s.%s'`, dbName, cName),
			}
			AssertEqualCommandError(t, expected, err)
		})
	}
}

func TestDiffCollectionName(t *testing.T) {
	t.Parallel()

	testcases := map[string]string{
		"ReservedPrefix": "_ferretdb_xxx",
		"NonUTF-8":       string([]byte{0xff, 0xfe, 0xfd}),
	}

	t.Run("CreateCollection", func(t *testing.T) {
		for name, cName := range testcases {
			name, cName := name, cName
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ctx, collection := setup.Setup(t)

				err := collection.Database().CreateCollection(ctx, cName)

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				expected := mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: fmt.Sprintf(`Invalid collection name: %s`, cName),
				}
				AssertEqualCommandError(t, expected, err)
			})
		}
	})

	t.Run("RenameCollection", func(t *testing.T) {
		for name, toName := range testcases {
			name, toName := name, toName
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ctx, collection := setup.Setup(t)

				fromName := testutil.CollectionName(t)
				err := collection.Database().CreateCollection(ctx, fromName)
				require.NoError(t, err)

				dbName := collection.Database().Name()
				command := bson.D{
					{"renameCollection", dbName + "." + fromName},
					{"to", dbName + "." + toName},
				}

				err = collection.Database().Client().Database("admin").RunCommand(ctx, command).Err()

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				expected := mongo.CommandError{
					Name:    "IllegalOperation",
					Code:    20,
					Message: fmt.Sprintf(`error with target namespace: Invalid collection name: %s`, toName),
				}
				AssertEqualCommandError(t, expected, err)
			})
		}
	})
}

func TestDiffNullStrings(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	_, err := collection.InsertOne(ctx, bson.D{
		{"_id", "document"},
		{"a", string([]byte{0})},
	})

	if setup.IsMongoDB(t) || setup.IsSQLite(t) {
		require.NoError(t, err)
		return
	}

	require.Error(t, err)
	assert.Regexp(t, "^.* unsupported Unicode escape sequence .*$", err.Error())
}

func TestDiffNegativeZero(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		insert bson.D
		update bson.D
		filter bson.D
	}{
		"Insert": {
			insert: bson.D{{"_id", "1"}, {"v", math.Copysign(0.0, -1)}},
			filter: bson.D{{"_id", "1"}},
		},
		"UpdateZeroMulNegative": {
			insert: bson.D{{"_id", "zero"}, {"v", int32(0)}},
			update: bson.D{{"$mul", bson.D{{"v", float64(-1)}}}},
			filter: bson.D{{"_id", "zero"}},
		},
		"UpdateNegativeMulZero": {
			insert: bson.D{{"_id", "negative"}, {"v", int64(-1)}},
			update: bson.D{{"$mul", bson.D{{"v", float64(0)}}}},
			filter: bson.D{{"_id", "negative"}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := collection.InsertOne(ctx, tc.insert)
			require.NoError(t, err)

			if tc.update != nil {
				_, err = collection.UpdateOne(ctx, tc.filter, tc.update)
				require.NoError(t, err)
			}

			var res bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&res)
			require.NoError(t, err)

			doc := ConvertDocument(t, res)
			v, _ := doc.Get("v")
			actual, ok := v.(float64)
			require.True(t, ok)
			require.Equal(t, 0.0, actual)

			if setup.IsMongoDB(t) {
				require.Equal(t, math.Signbit(math.Copysign(0.0, -1)), math.Signbit(actual))
				return
			}

			require.Equal(t, math.Signbit(math.Copysign(0.0, +1)), math.Signbit(actual))
		})
	}
}

func TestDiffUpdateProduceInfinity(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)
	_, err := collection.InsertOne(ctx, bson.D{{"_id", "number"}, {"v", int32(42)}})
	require.NoError(t, err)

	_, err = collection.UpdateOne(ctx, bson.D{{"_id", "number"}}, bson.D{{"$mul", bson.D{{"v", math.MaxFloat64}}}})

	if setup.IsMongoDB(t) {
		require.NoError(t, err)
		return
	}

	expected := mongo.CommandError{
		Code: 2,
		Name: "BadValue",
		Message: `update produces invalid value: { "v": +Inf }` +
			` (update operations that produce infinity values are not allowed)`,
	}
	AssertEqualCommandError(t, expected, err)
}

func TestDiffDocumentValidation(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars)

	t.Run("Insert", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct { //nolint:vet // use only for testing
			doc bson.D

			err error
		}{
			"DollarSign": {
				doc: bson.D{{"$foo", "bar"}},
				err: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Index:   0,
					Code:    2,
					Message: `invalid key: "$foo" (key must not start with '$' sign)`,
				}}},
			},
			"DotSign": {
				doc: bson.D{{"foo.bar", "baz"}},
				err: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Index:   0,
					Code:    2,
					Message: `invalid key: "foo.bar" (key must not contain '.' sign)`,
				}}},
			},
			"Infinity": {
				doc: bson.D{{"foo", math.Inf(1)}},
				err: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Code:    2,
					Message: `invalid value: { "foo": +Inf } (infinity values are not allowed)`,
				}}},
			},
			"NegativeInfinity": {
				doc: bson.D{{"foo", math.Inf(-1)}},
				err: mongo.WriteException{WriteErrors: []mongo.WriteError{{
					Code:    2,
					Message: `invalid value: { "foo": -Inf } (infinity values are not allowed)`,
				}}},
			},
			"NaN": {
				doc: bson.D{{"_id", "nan"}, {"foo", math.NaN()}},
				err: mongo.CommandError{
					Code: 2,
					Name: "BadValue",
					Message: `wire.OpMsg.Document: validation failed for { insert: "TestDiffDocumentValidation", ordered: true, ` +
						`$db: "TestDiffDocumentValidation", documents: [ { _id: "nan", foo: nan.0 } ] } with: NaN is not supported`,
				},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := collection.InsertOne(ctx, tc.doc)

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				assert.Equal(t, tc.err, UnsetRaw(t, err))
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			doc bson.D
			err mongo.WriteError
		}{
			"DotSign": {
				doc: bson.D{{"$set", bson.D{{"foo", bson.D{{"bar.baz", "qaz"}}}}}},
				err: mongo.WriteError{
					Code:    2,
					Message: `invalid key: "bar.baz" (key must not contain '.' sign)`,
				},
			},
		} {
			name, tc := name, tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := collection.UpdateOne(ctx, bson.D{}, tc.doc)

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				AssertEqualWriteError(t, tc.err, err)
			})
		}
	})
}

func TestDiffDBStatsScale(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	testCases := map[string]struct {
		scale               any
		expectedMongoDBErr  string
		expectedFerretDBErr mongo.CommandError
	}{
		"Zero": {
			scale:              int32(0),
			expectedMongoDBErr: "scale has to be > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "Location51024",
				Code:    51024,
				Message: "BSON field 'scale' value must be >= 1, actual value '0'",
			},
		},
		"Negative": {
			scale:              int32(-100),
			expectedMongoDBErr: "scale has to be > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "Location51024",
				Code:    51024,
				Message: "BSON field 'scale' value must be >= 1, actual value '-100'",
			},
		},
		"MinFloat": {
			scale:              -math.MaxFloat64,
			expectedMongoDBErr: "scale has to be > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "Location51024",
				Code:    51024,
				Message: "BSON field 'scale' value must be >= 1, actual value '-9223372036854775808'",
			},
		},
		"String": {
			scale:              "1",
			expectedMongoDBErr: "scale has to be a number > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "TypeMismatch",
				Code:    14,
				Message: "BSON field 'dbStats.scale' is the wrong type 'string', expected types '[long, int, decimal, double]'",
			},
		},
		"Object": {
			scale:              bson.D{{"a", 1}},
			expectedMongoDBErr: "scale has to be a number > 0",
			expectedFerretDBErr: mongo.CommandError{
				Name:    "TypeMismatch",
				Code:    14,
				Message: "BSON field 'dbStats.scale' is the wrong type 'object', expected types '[long, int, decimal, double]'",
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := collection.Database().RunCommand(ctx, bson.D{{"dbStats", int32(1)}, {"scale", tc.scale}}).Err()
			require.Error(t, err)

			if setup.IsMongoDB(t) {
				expected := mongo.CommandError{
					Name:    "",
					Code:    0,
					Message: tc.expectedMongoDBErr,
				}
				AssertEqualCommandError(t, expected, err)
				return
			}

			AssertEqualCommandError(t, tc.expectedFerretDBErr, err)
		})
	}
}

func TestDiffErrorMessages(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})
	ctx, db := s.Ctx, s.Collection.Database()

	err := db.RunCommand(ctx, bson.D{{"getParameter", bson.D{{"allParameters", "1"}}}}).Err()

	if setup.IsMongoDB(t) {
		expected := mongo.CommandError{
			Code: 14,
			Name: "TypeMismatch",
			Message: "BSON field 'getParameter.allParameters' is the wrong type 'string', " +
				"expected types '[bool, long, int, decimal, double']",
		}
		AssertEqualCommandError(t, expected, err)

		return
	}

	expected := mongo.CommandError{
		Code:    14,
		Name:    "TypeMismatch",
		Message: "BSON field 'allParameters' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
	}
	AssertEqualCommandError(t, expected, err)
}
