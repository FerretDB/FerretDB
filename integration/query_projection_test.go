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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryProjectionErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		filter     bson.D // required
		projection any    // required

		err        *mongo.CommandError // required, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"EmptyKey": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"", true}},
			err: &mongo.CommandError{
				Code:    40352,
				Name:    "Location40352",
				Message: "FieldPath cannot be constructed with empty string",
			},
		},
		"EmptyPath": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"v..d", true}},
			err: &mongo.CommandError{
				Code:    15998,
				Name:    "Location15998",
				Message: "FieldPath field names may not be empty strings.",
			},
		},
		"ExcludeInclude": {
			filter:     bson.D{},
			projection: bson.D{{"foo", false}, {"_id", false}, {"bar", true}},
			err: &mongo.CommandError{
				Code:    31253,
				Name:    "Location31253",
				Message: "Cannot do inclusion on field bar in exclusion projection",
			},
		},
		"IncludeExclude": {
			filter:     bson.D{},
			projection: bson.D{{"foo", true}, {"_id", false}, {"bar", false}},
			err: &mongo.CommandError{
				Code:    31254,
				Name:    "Location31254",
				Message: "Cannot do exclusion on field bar in inclusion projection",
			},
		},
		"PositionalOperatorMultiple": {
			filter:     bson.D{{"_id", "array-numbers-asc"}},
			projection: bson.D{{"v.$.foo.$", true}},
			err: &mongo.CommandError{
				Code: 31394,
				Name: "Location31394",
				Message: "As of 4.4, it's illegal to specify positional operator " +
					"in the middle of a path.Positional projection may only be " +
					"used at the end, for example: a.b.$. If the query previously " +
					"used a form like a.b.$.d, remove the parts following the '$' and " +
					"the results will be equivalent.",
			},
			altMessage: "Positional projection may only be used at the end, " +
				"for example: a.b.$. If the query previously used a form " +
				"like a.b.$.d, remove the parts following the '$' and " +
				"the results will be equivalent.",
		},
		"PositionalOperatorMiddle": {
			filter:     bson.D{{"_id", "array-numbers-asc"}},
			projection: bson.D{{"v.$.foo", true}},
			err: &mongo.CommandError{
				Code: 31394,
				Name: "Location31394",
				Message: "As of 4.4, it's illegal to specify positional operator " +
					"in the middle of a path.Positional projection may only be " +
					"used at the end, for example: a.b.$. If the query previously " +
					"used a form like a.b.$.d, remove the parts following the '$' and " +
					"the results will be equivalent.",
			},
			altMessage: "Positional projection may only be used at the end, " +
				"for example: a.b.$. If the query previously used a form " +
				"like a.b.$.d, remove the parts following the '$' and " +
				"the results will be equivalent.",
		},
		"PositionalOperatorWrongLocations": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"$.v.$.foo", true}},
			err: &mongo.CommandError{
				Code: 31394,
				Name: "Location31394",
				Message: "As of 4.4, it's illegal to specify positional operator " +
					"in the middle of a path.Positional projection may only be " +
					"used at the end, for example: a.b.$. If the query previously " +
					"used a form like a.b.$.d, remove the parts following the '$' and " +
					"the results will be equivalent.",
			},
			altMessage: "Positional projection may only be " +
				"used at the end, for example: a.b.$. If the query previously " +
				"used a form like a.b.$.d, remove the parts following the '$' and " +
				"the results will be equivalent.",
		},
		"PositionalOperatorEmptyFilter": {
			filter:     bson.D{},
			projection: bson.D{{"v.$", true}},
			err: &mongo.CommandError{
				Code: 51246,
				Name: "Location51246",
				Message: "Executor error during find command :: caused by :: " +
					"positional operator '.$' couldn't find a matching element in the array",
			},
		},
		"PositionalOperatorEmptyArrayID": {
			filter:     bson.D{{"_id", "array-empty"}},
			projection: bson.D{{"v.$", true}},
			err: &mongo.CommandError{
				Code:    51246,
				Name:    "Location51246",
				Message: "Executor error during find command :: caused by :: positional operator '.$' couldn't find a matching element in the array",
			},
		},
		"PositionalOperatorFilterMissingPath": {
			// it returns error only if collection contains a doc that matches the filter
			// and the filter does not contain positional operator path,
			// e.g. missing {v: <val>} in the filter.
			filter:     bson.D{{"_id", "array"}},
			projection: bson.D{{"v.$", true}},
			err: &mongo.CommandError{
				Code:    51246,
				Name:    "Location51246",
				Message: "Executor error during find command :: caused by :: positional operator '.$' couldn't find a matching element in the array",
			},
		},
		"PositionalOperatorMismatch": {
			// positional projection only handles one array at the suffix,
			// path prefixes cannot contain array.
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"v.foo.$", true}},
			err: &mongo.CommandError{
				Code:    51247,
				Name:    "Location51247",
				Message: "Executor error during find command :: caused by :: positional operator '.$' element mismatch",
			},
		},
		"PositionalOperatorEmptyPath": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"v..$", true}},
			err: &mongo.CommandError{
				Code:    40353,
				Name:    "Location40353",
				Message: "FieldPath must not end with a '.'.",
			},
		},
		"PositionalOperatorDollarKey": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"$", true}},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PositionalOperatorDollarInKey": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"$v", true}},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PositionalOperatorDollarPrefix": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"$.foo", true}},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PositionalOperatorDotNotationDollarInKey": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"v.$foo", true}},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PositionalOperatorPrefixSuffix": {
			filter:     bson.D{{"_id", "array-numbers-asc"}},
			projection: bson.D{{"$.foo.$", true}},
			err: &mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PositionalOperatorExclusion": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"v.$", false}},
			err: &mongo.CommandError{
				Code:    31395,
				Name:    "Location31395",
				Message: "positional projection cannot be used with exclusion",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.filter, "filter should be set")
			require.NotNil(t, tc.projection, "projection should be set")
			require.NotNil(t, tc.err, "err should be set")

			res, err := collection.Find(ctx, tc.filter, options.Find().SetProjection(tc.projection))

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}
