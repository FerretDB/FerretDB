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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func TestAggregateAddFieldsErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		pipeline bson.A // required, aggregation pipeline stages

		err              *mongo.CommandError // required
		altMessage       string              // optional, alternative error message
		failsForFerretDB string
	}{
		"NotDocument": {
			pipeline: bson.A{
				bson.D{{"$addFields", "not-document"}},
			},
			err: &mongo.CommandError{
				Code:    40272,
				Name:    "Location40272",
				Message: "$addFields specification stage must be an object, got string",
			},
			altMessage: "$addFields specification stage must be an object",
		},
		"InvalidFieldPath": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"$foo", "v"}}}},
			},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "Invalid $addFields :: caused by :: FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.pipeline, "pipeline must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			_, err := collection.Aggregate(ctx, tc.pipeline)

			if tc.altMessage != "" {
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			AssertEqualCommandError(t, *tc.err, err)
		})
	}
}

func TestAggregateGroupSumDecimalDouble(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1056")

	ctx, collection := setup.Setup(t)

	_, err := collection.InsertMany(ctx, bson.A{
		bson.D{{"_id", "decimal128"}, {"v", must.NotFail(primitive.ParseDecimal128("42.1"))}},
		bson.D{{"_id", "double"}, {"v", float64(42.1)}},
	})
	require.NoError(t, err)

	cursor, err := collection.Aggregate(ctx, bson.A{
		bson.D{{"$group", bson.D{{"_id", nil}, {"sum", bson.D{{"$sum", "$v"}}}}}},
	})
	require.NoError(t, err)

	var res bson.A
	require.NoError(t, cursor.All(ctx, &res))

	expected := bson.A{bson.D{
		{"_id", nil},
		{"sum", primitive.NewDecimal128(3459220962935157325, 6906845732440572485)}, // 84.20000000000000142108547152020037
	}}

	assert.Equal(t, expected, res)
}

func TestAggregateGroupErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		pipeline bson.A // required, aggregation pipeline stages

		err              *mongo.CommandError // required, expected error from MongoDB
		altMessage       string              // optional, alternative error message for FerretDB, ignored if empty
		skip             string              // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1086
		failsForFerretDB string
	}{
		"UnaryOperatorSum": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"sum", bson.D{{"$sum", bson.A{}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    40237,
				Name:    "Location40237",
				Message: "The $sum accumulator is a unary operator",
			},
			altMessage:       "The $sum accumulator is a unary operator",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/389",
		},
		"TypeEmpty": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"v", bson.D{}}}}},
			},
			err: &mongo.CommandError{
				Code:    40234,
				Name:    "Location40234",
				Message: "The field 'v' must be an accumulator object",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/389",
		},
		"TwoOperators": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$type", int32(42)}, {"$op", int32(42)}}}}}},
			},
			err: &mongo.CommandError{
				Code:    15983,
				Name:    "Location15983",
				Message: "An object representing an expression must have exactly one field: { $type: 42, $op: 42 }",
			},
			altMessage:       "An object representing an expression must have exactly one field",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"TypeInvalidLen": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$type", bson.A{"foo", "bar"}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Expression $type takes exactly 1 arguments. 2 were passed in.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"NonExistentOperator": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$non-existent", "foo"}}}}}},
			},
			err: &mongo.CommandError{
				Code:    168,
				Name:    "InvalidPipelineOperator",
				Message: "Unrecognized expression '$non-existent'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"SumEmptyExpression": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"$sum", "$"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    16872,
				Name:    "Location16872",
				Message: "'$' by itself is not a valid FieldPath",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"SumEmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"$sum", "$$"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "empty variable names are not allowed",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"SumDollarVariable": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"$sum", "$$$"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "'$' starts with an invalid character for a user variable name",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"RecursiveNonExistentOperator": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$type", bson.D{{"$non-existent", "foo"}}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    168,
				Name:    "InvalidPipelineOperator",
				Message: "Unrecognized expression '$non-existent'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"IDExpressionDuplicateFields": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"v", "$v"},
						{"v", "$non-existent"},
					}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    16406,
				Name:    "Location16406",
				Message: "duplicate field name specified in object literal: { v: \"$v\", v: \"$non-existent\" }",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"IDExpressionEmptyPath": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", "$"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    16872,
				Name:    "Location16872",
				Message: "'$' by itself is not a valid FieldPath",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"IDExpressionEmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", "$$"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "empty variable names are not allowed",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"IDExpressionInvalidVariable$": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", "$$$"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "'$' starts with an invalid character for a user variable name",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"IDExpressionInvalidVariable$s": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", "$$$s"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "'$s' starts with an invalid character for a user variable name",
			},
			altMessage:       "'$' starts with an invalid character for a user variable name",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"IDExpressionNonExistingVariable": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", "$$s"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    17276,
				Name:    "Location17276",
				Message: "Use of undefined variable: s",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2275",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			if tc.skip != "" {
				tt.Skip(tc.skip)
			}

			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.pipeline, "pipeline must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			res, err := collection.Aggregate(ctx, tc.pipeline)

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestAggregateProjectErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		pipeline bson.A // required, aggregation pipeline stages

		err              *mongo.CommandError // required
		altMessage       string              // optional, alternative error message
		skip             string              // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1086
		failsForFerretDB string
	}{
		"EmptyPipeline": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{}}},
			},
			err: &mongo.CommandError{
				Code:    51272,
				Name:    "Location51272",
				Message: "Invalid $project :: caused by :: projection specification must have at least one field",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"EmptyProjectionField": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v", bson.D{}}}}},
			},
			err: &mongo.CommandError{
				Code:    51270,
				Name:    "Location51270",
				Message: "Invalid $project :: caused by :: An empty sub-projection is not a valid value. Found empty object at path",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2633",
		},
		"EmptyKey": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"", true}}}},
			},
			err: &mongo.CommandError{
				Code: 40352,
				Name: "Location40352",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath cannot be constructed with empty string",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"EmptyPath": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v..d", true}}}},
			},
			err: &mongo.CommandError{
				Code:    15998,
				Name:    "Location15998",
				Message: "Invalid $project :: caused by :: FieldPath field names may not be empty strings.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"ExcludeInclude": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"foo", false}, {"bar", true}}}},
			},
			err: &mongo.CommandError{
				Code:    31253,
				Name:    "Location31253",
				Message: "Invalid $project :: caused by :: Cannot do inclusion on field bar in exclusion projection",
			},
			altMessage:       "Cannot do inclusion on field bar in exclusion projection",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"IncludeExclude": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"foo", true}, {"bar", false}}}},
			},
			err: &mongo.CommandError{
				Code:    31254,
				Name:    "Location31254",
				Message: "Invalid $project :: caused by :: Cannot do exclusion on field bar in inclusion projection",
			},
			altMessage:       "Cannot do exclusion on field bar in inclusion projection",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorMultiple": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$.foo.$", true}}}},
			},
			err: &mongo.CommandError{
				Code: 31324,
				Name: "Location31324",
				Message: "Invalid $project :: caused by :: " +
					"Cannot use positional projection in aggregation projection",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorMiddle": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$.foo", true}}}},
			},
			err: &mongo.CommandError{
				Code: 31394,
				Name: "Location31394",
				Message: "Invalid $project :: caused by :: " +
					"As of 4.4, it's illegal to specify positional operator " +
					"in the middle of a path.Positional projection may only be " +
					"used at the end, for example: a.b.$. If the query previously " +
					"used a form like a.b.$.d, remove the parts following the '$' and " +
					"the results will be equivalent.",
			},
			altMessage: "Invalid $project :: caused by :: " +
				"Positional projection may only be used at the end, " +
				"for example: a.b.$. If the query previously used a form " +
				"like a.b.$.d, remove the parts following the '$' and " +
				"the results will be equivalent.",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorWrongLocations": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$.v.$.foo", true}}}},
			},
			err: &mongo.CommandError{
				Code: 31394,
				Name: "Location31394",
				Message: "Invalid $project :: caused by :: " +
					"As of 4.4, it's illegal to specify positional operator " +
					"in the middle of a path.Positional projection may only be " +
					"used at the end, for example: a.b.$. If the query previously " +
					"used a form like a.b.$.d, remove the parts following the '$' and " +
					"the results will be equivalent.",
			},
			altMessage: "Invalid $project :: caused by :: " +
				"Positional projection may only be used at the end, " +
				"for example: a.b.$. If the query previously used a form " +
				"like a.b.$.d, remove the parts following the '$' and " +
				"the results will be equivalent.",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorEmptyPath": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v..$", true}}}},
			},
			err: &mongo.CommandError{
				Code: 31324,
				Name: "Location31324",
				Message: "Invalid $project :: caused by :: " +
					"Cannot use positional projection in aggregation projection",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorDollarKey": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$", true}}}},
			},
			err: &mongo.CommandError{
				Code: 16410,
				Name: "Location16410",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorDollarInKey": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$v", true}}}},
			},
			err: &mongo.CommandError{
				Code: 16410,
				Name: "Location16410",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorDollarPrefix": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$.foo", true}}}},
			},
			err: &mongo.CommandError{
				Code: 16410,
				Name: "Location16410",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorDotDollarInKey": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$foo", true}}}},
			},
			err: &mongo.CommandError{
				Code: 16410,
				Name: "Location16410",
				Message: "Invalid $project :: caused by :: " +
					"FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorPrefixSuffix": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"$.v.$", true}}}},
			},
			err: &mongo.CommandError{
				Code: 31324,
				Name: "Location31324",
				Message: "Invalid $project :: caused by :: " +
					"Cannot use positional projection in aggregation projection",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"PositionalOperatorExclusion": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$", false}}}},
			},
			err: &mongo.CommandError{
				Code: 31324,
				Name: "Location31324",
				Message: "Invalid $project :: caused by :: " +
					"Cannot use positional projection in aggregation projection",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"ProjectPositionalOperator": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v.$", true}}}},
			},
			err: &mongo.CommandError{
				Code:    31324,
				Name:    "Location31324",
				Message: "Invalid $project :: caused by :: Cannot use positional projection in aggregation projection",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"ProjectTypeEmpty": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v", bson.D{}}}}},
			},
			err: &mongo.CommandError{
				Code:    51270,
				Name:    "Location51270",
				Message: "Invalid $project :: caused by :: An empty sub-projection is not a valid value." + " Found empty object at path",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"ProjectTwoOperators": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v", bson.D{{"$type", int32(42)}, {"$op", int32(42)}}}}}},
			},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "Invalid $project :: caused by :: FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"ProjectTypeInvalidLen": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v", bson.D{{"$type", bson.A{"foo", "bar"}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    16020,
				Name:    "Location16020",
				Message: "Invalid $project :: caused by :: Expression $type takes exactly 1 arguments. 2 were passed in.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"ProjectNonExistentOperator": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v", bson.D{{"$non-existent", "foo"}}}}}},
			},
			altMessage: "Invalid $project :: caused by :: Unrecognized expression '$non-existent'",
			err: &mongo.CommandError{
				Code:    31325,
				Name:    "Location31325",
				Message: "Invalid $project :: caused by :: Unknown expression $non-existent",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"ProjectRecursiveNonExistentOperator": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v", bson.D{{"$type", bson.D{{"$non-existent", "foo"}}}}}}}},
			},
			err: &mongo.CommandError{
				Code:    168,
				Name:    "InvalidPipelineOperator",
				Message: "Invalid $project :: caused by :: Unrecognized expression '$non-existent'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"SumEmptyExpression": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    16872,
				Name:    "Location16872",
				Message: "Invalid $project :: caused by :: '$' by itself is not a valid FieldPath",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"SumEmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$$"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Invalid $project :: caused by :: empty variable names are not allowed",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"SumDollarVariable": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$$$"}}},
				}}},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Invalid $project :: caused by :: '$' starts with an invalid character for a user variable name",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			if tc.skip != "" {
				tt.Skip(tc.skip)
			}

			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.pipeline, "pipeline must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			_, err := collection.Aggregate(ctx, tc.pipeline)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestAggregateProject(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		pipeline bson.A // required, aggregation pipeline stages

		res []bson.D // required, expected response
	}{
		"IDFalseValueTrue": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"_id", "int32"}}}},
				bson.D{{"$project", bson.D{{"_id", false}, {"v", true}}}},
			},
			res: []bson.D{{{"v", int32(42)}}},
		},
		"ValueTrueIDFalse": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"_id", "int32"}}}},
				bson.D{{"$project", bson.D{{"v", true}, {"_id", false}}}},
			},
			res: []bson.D{{{"v", int32(42)}}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, tc.pipeline, "pipeline must not be nil")
			require.NotNil(t, tc.res, "res must not be nil")

			cursor, err := collection.Aggregate(ctx, tc.pipeline)
			require.NoError(t, err)
			defer cursor.Close(ctx)

			var res []bson.D
			err = cursor.All(ctx, &res)
			require.NoError(t, err)
			require.Equal(t, tc.res, res)
		})
	}
}

func TestAggregateSetErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		pipeline bson.A // required, aggregation pipeline stages

		err              *mongo.CommandError // required
		altMessage       string              // optional, alternative error message
		failsForFerretDB string
	}{
		"NotDocument": {
			pipeline: bson.A{
				bson.D{{"$set", "not-document"}},
			},
			err: &mongo.CommandError{
				Code:    40272,
				Name:    "Location40272",
				Message: "$set specification stage must be an object, got string",
			},
			altMessage: "$addFields specification stage must be an object",
		},
		"InvalidFieldPath": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"$foo", "v"}}}},
			},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "Invalid $set :: caused by :: FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.pipeline, "pipeline must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			_, err := collection.Aggregate(ctx, tc.pipeline)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestAggregateUnsetErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		pipeline bson.A // required, aggregation pipeline stages

		err              *mongo.CommandError // required
		altMessage       string              // optional, alternative error message
		failsForFerretDB string
	}{
		"EmptyString": {
			pipeline: bson.A{
				bson.D{{"$unset", ""}},
			},
			err: &mongo.CommandError{
				Code:    40352,
				Name:    "Location40352",
				Message: "Invalid $unset :: caused by :: FieldPath cannot be constructed with empty string",
			},
			altMessage: "FieldPath cannot be constructed with empty string",
		},
		"InvalidType": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.D{}}},
			},
			err: &mongo.CommandError{
				Code:    31002,
				Name:    "Location31002",
				Message: "$unset specification must be a string or an array",
			},
		},
		"PathEmptyKey": {
			pipeline: bson.A{
				bson.D{{"$unset", "v..foo"}},
			},
			err: &mongo.CommandError{
				Code:    15998,
				Name:    "Location15998",
				Message: "Invalid $unset :: caused by :: FieldPath field names may not be empty strings.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"PathEmptySuffixKey": {
			pipeline: bson.A{
				bson.D{{"$unset", "v."}},
			},
			err: &mongo.CommandError{
				Code:    40353,
				Name:    "Location40353",
				Message: "Invalid $unset :: caused by :: FieldPath must not end with a '.'.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"PathEmptyPrefixKey": {
			pipeline: bson.A{
				bson.D{{"$unset", ".v"}},
			},
			err: &mongo.CommandError{
				Code:    15998,
				Name:    "Location15998",
				Message: "Invalid $unset :: caused by :: FieldPath field names may not be empty strings.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"PathDollarPrefix": {
			pipeline: bson.A{
				bson.D{{"$unset", "$v"}},
			},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "Invalid $unset :: caused by :: FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
			altMessage: "FieldPath field names may not start with '$'",
		},
		"ArrayEmpty": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{}}},
			},
			err: &mongo.CommandError{
				Code:    31119,
				Name:    "Location31119",
				Message: "$unset specification must be a string or an array with at least one field",
			},
			altMessage: "$unset specification must be a string or an array with at least one field",
		},
		"ArrayInvalidType": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"field1", 1}}},
			},
			err: &mongo.CommandError{
				Code:    31120,
				Name:    "Location31120",
				Message: "$unset specification must be a string or an array containing only string values",
			},
		},
		"ArrayEmptyString": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{""}}},
			},
			err: &mongo.CommandError{
				Code:    40352,
				Name:    "Location40352",
				Message: "Invalid $unset :: caused by :: FieldPath cannot be constructed with empty string",
			},
			altMessage: "FieldPath cannot be constructed with empty string",
		},
		"ArrayPathDuplicate": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v", "v"}}},
			},
			err: &mongo.CommandError{
				Code:    31250,
				Name:    "Location31250",
				Message: "Invalid $unset :: caused by :: Path collision at v",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"ArrayPathOverwrites": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v", "v.foo"}}},
			},
			err: &mongo.CommandError{
				Code:    31249,
				Name:    "Location31249",
				Message: "Invalid $unset :: caused by :: Path collision at v.foo remaining portion foo",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"ArrayPathOverwritesRemaining": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v", "v.foo.bar"}}},
			},
			err: &mongo.CommandError{
				Code:    31249,
				Name:    "Location31249",
				Message: "Invalid $unset :: caused by :: Path collision at v.foo.bar remaining portion foo.bar",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"ArrayPathCollision": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v.foo", "v"}}},
			},
			err: &mongo.CommandError{
				Code:    31250,
				Name:    "Location31250",
				Message: "Invalid $unset :: caused by :: Path collision at v",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"ArrayPathEmptyKey": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v..foo"}}},
			},
			err: &mongo.CommandError{
				Code:    15998,
				Name:    "Location15998",
				Message: "Invalid $unset :: caused by :: FieldPath field names may not be empty strings.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"ArrayPathEmptySuffixKey": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v."}}},
			},
			err: &mongo.CommandError{
				Code:    40353,
				Name:    "Location40353",
				Message: "Invalid $unset :: caused by :: FieldPath must not end with a '.'.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"ArrayPathEmptyPrefixKey": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{".v"}}},
			},
			err: &mongo.CommandError{
				Code:    15998,
				Name:    "Location15998",
				Message: "Invalid $unset :: caused by :: FieldPath field names may not be empty strings.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"ArrayPathDollarPrefix": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"$v"}}},
			},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "Invalid $unset :: caused by :: FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
			altMessage: "FieldPath field names may not start with '$'",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.pipeline, "pipeline must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			_, err := collection.Aggregate(ctx, tc.pipeline)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestAggregateSortErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		pipeline bson.A // required, aggregation pipeline stages

		err              *mongo.CommandError // required
		altMessage       string              // optional, alternative error message
		failsForFerretDB string
	}{
		"DotNotationMissingField": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v..foo", 1},
			}}}},
			err: &mongo.CommandError{
				Code:    15998,
				Name:    "Location15998",
				Message: "FieldPath field names may not be empty strings.",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.pipeline, "pipeline must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			_, err := collection.Aggregate(ctx, tc.pipeline)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestAggregateCommandMaxTimeMSErrors(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		command bson.D // required, command to run

		err              *mongo.CommandError // required, expected error from MongoDB
		altMessage       string              // optional, alternative error message for FerretDB, ignored if empty
		failsForFerretDB string
	}{
		"NegativeLong": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", int64(-1)},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'maxTimeMS' value must be >= 0, actual value '-1'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"MaxLong": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", math.MaxInt64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "9223372036854775807 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage:       "9223372036854775807 value for maxTimeMS is out of range",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"Double": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", 1000.5},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS has non-integral value",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"NegativeDouble": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", -14245345234123245.55},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'maxTimeMS' value must be >= 0, actual value '-14245345234123246'",
			},
			altMessage:       "BSON field 'maxTimeMS' value must be >= 0, actual value '-1.424534523412325e+16'",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"BigDouble": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", math.MaxFloat64},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "9223372036854775807 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage:       "1.797693134862316e+308 value for maxTimeMS is out of range",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"BigNegativeDouble": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", -math.MaxFloat64},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'maxTimeMS' value must be >= 0, actual value '-9223372036854775808'",
			},
			altMessage:       "BSON field 'maxTimeMS' value must be >= 0, actual value '-1.797693134862316e+308'",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"NegativeInt32": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", -1123123},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'maxTimeMS' value must be >= 0, actual value '-1123123'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"MaxIntPlus": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", math.MaxInt32 + 1},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "2147483648 value for maxTimeMS is out of range " + shareddata.Int32Interval,
			},
			altMessage:       "2147483648 value for maxTimeMS is out of range",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"Null": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", nil},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "maxTimeMS must be a number",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"String": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", "string"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'aggregate.maxTimeMS' is the wrong type 'string', expected types '[long, int, decimal, double']",
			},
			altMessage:       "BSON field 'aggregate.maxTimeMS' is the wrong type 'string', expected types '[long, int, decimal, double]'",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"Array": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", bson.A{int32(42), "foo", nil}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'aggregate.maxTimeMS' is the wrong type 'array', expected types '[long, int, decimal, double']",
			},
			altMessage:       "BSON field 'aggregate.maxTimeMS' is the wrong type 'array', expected types '[long, int, decimal, double]'",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"Document": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", bson.D{{"foo", int32(42)}}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'aggregate.maxTimeMS' is the wrong type 'object', expected types '[long, int, decimal, double']",
			},
			altMessage:       "BSON field 'aggregate.maxTimeMS' is the wrong type 'object', expected types '[long, int, decimal, double]'",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			require.NotNil(t, tc.err, "err must not be nil")

			var res bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
			require.Nil(t, res)
		})
	}
}

func TestAggregateCommandCursor(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	// the number of documents is set above the default batchSize of 101
	// for testing unset batchSize returning default batchSize
	arr := GenerateDocuments(0, 110)
	_, err := collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		pipeline any // optional, defaults to bson.A{}
		cursor   any // optional, nil to leave cursor unset

		firstBatch       primitive.A         // optional, expected firstBatch
		err              *mongo.CommandError // optional, expected error from MongoDB
		altMessage       string              // optional, alternative error message for FerretDB, ignored if empty
		failsForFerretDB string
	}{
		"Int": {
			cursor:     bson.D{{"batchSize", 1}},
			firstBatch: arr[:1],
		},
		"Long": {
			cursor:     bson.D{{"batchSize", int64(2)}},
			firstBatch: arr[:2],
		},
		"LongZero": {
			cursor:     bson.D{{"batchSize", int64(0)}},
			firstBatch: bson.A{},
		},
		"LongNegative": {
			cursor: bson.D{{"batchSize", int64(-1)}},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
			altMessage:       "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
		},
		"DoubleZero": {
			cursor:     bson.D{{"batchSize", float64(0)}},
			firstBatch: bson.A{},
		},
		"DoubleNegative": {
			cursor: bson.D{{"batchSize", -1.1}},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: "BSON field 'batchSize' value must be >= 0, actual value '-1'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
		},
		"DoubleFloor": {
			cursor:     bson.D{{"batchSize", 1.9}},
			firstBatch: arr[:1],
		},
		"Bool": {
			cursor:     bson.D{{"batchSize", true}},
			firstBatch: arr[:1],
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'cursor.batchSize' is the wrong type 'bool', expected types '[long, int, decimal, double']",
			},
			altMessage: "BSON field 'cursor.batchSize' is the wrong type 'bool', expected type 'number'",
		},
		"Unset": {
			cursor:     nil,
			firstBatch: arr[:101],
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "The 'cursor' option is required, except for aggregate with the explain argument",
			},
		},
		"Empty": {
			cursor:     bson.D{},
			firstBatch: arr[:101],
		},
		"String": {
			cursor:     "invalid",
			firstBatch: arr[:101],
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "cursor field must be missing or an object",
			},
			altMessage: "BSON field 'cursor' is the wrong type 'string', expected type 'object'",
		},
		"LargeBatchSize": {
			cursor:     bson.D{{"batchSize", 102}},
			firstBatch: arr[:102],
		},
		"LargeBatchSizeMatch": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"_id", bson.D{{"$in", bson.A{0, 1, 2, 3, 4, 5}}}}}}},
			},
			cursor:     bson.D{{"batchSize", 102}},
			firstBatch: arr[:6],
		},
	} {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testing.TB = tt

			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			var pipeline any = bson.A{}
			if tc.pipeline != nil {
				pipeline = tc.pipeline
			}

			var rest bson.D
			if tc.cursor != nil {
				rest = append(rest, bson.E{Key: "cursor", Value: tc.cursor})
			}

			command := append(
				bson.D{
					{"aggregate", collection.Name()},
					{"pipeline", pipeline},
				},
				rest...,
			)

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)
			if tc.err != nil {
				assert.Nil(t, res)
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

				return
			}

			require.NoError(t, err)

			v, ok := res.Map()["cursor"]
			require.True(t, ok)

			cursor, ok := v.(bson.D)
			require.True(t, ok)

			// do not check the value of cursor id, FerretDB has a different id
			cursorID := cursor.Map()["id"]
			assert.NotNil(t, cursorID)

			firstBatch, ok := cursor.Map()["firstBatch"]
			require.True(t, ok)
			require.Equal(t, tc.firstBatch, firstBatch)
		})
	}
}
