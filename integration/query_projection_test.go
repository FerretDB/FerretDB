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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryProjectionPositionalOperatorErrors(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct { //nolint:vet // used for testing only
		filter     bson.D             // required
		projection any                // required
		err        mongo.CommandError // required
		altMessage string             // optional
	}{
		"Multiple": {
			filter:     bson.D{{"_id", "array-numbers-asc"}},
			projection: bson.D{{"v.$.foo.$", true}},
			err: mongo.CommandError{
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
		"Middle": {
			filter:     bson.D{{"_id", "array-numbers-asc"}},
			projection: bson.D{{"v.$.foo", true}},
			err: mongo.CommandError{
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
		"WrongLocations": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"$.v.$.foo", true}},
			err: mongo.CommandError{
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
		"EmptyFilter": {
			filter:     bson.D{},
			projection: bson.D{{"v.$", true}},
			err: mongo.CommandError{
				Code: 51246,
				Name: "Location51246",
				Message: "Executor error during find command :: caused by :: " +
					"positional operator '.$' couldn't find a matching element in the array",
			},
		},
		"EmptyArrayID": {
			filter:     bson.D{{"_id", "array-empty"}},
			projection: bson.D{{"v.$", true}},
			err: mongo.CommandError{
				Code:    51246,
				Name:    "Location51246",
				Message: "Executor error during find command :: caused by :: positional operator '.$' couldn't find a matching element in the array",
			},
		},
		"FilterMissingPath": {
			// filter must contain filter for positional operator path v e.g. {v: <val>}
			filter:     bson.D{{"_id", "array"}},
			projection: bson.D{{"v.$", true}},
			err: mongo.CommandError{
				Code:    51246,
				Name:    "Location51246",
				Message: "Executor error during find command :: caused by :: positional operator '.$' couldn't find a matching element in the array",
			},
		},
		"Mismatch": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"v.foo.$", true}},
			err: mongo.CommandError{
				Code:    51247,
				Name:    "Location51247",
				Message: "Executor error during find command :: caused by :: positional operator '.$' element mismatch",
			},
		},
		"EmptyKey": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"", true}},
			err: mongo.CommandError{
				Code:    40352,
				Name:    "Location40352",
				Message: "FieldPath cannot be constructed with empty string",
			},
		},
		"EmptyPath": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"v..$", true}},
			err: mongo.CommandError{
				Code:    40353,
				Name:    "Location40353",
				Message: "FieldPath must not end with a '.'.",
			},
		},
		"DollarKey": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"$", true}},
			err: mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"DollarInKey": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"$v", true}},
			err: mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"DollarPrefix": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"$.foo", true}},
			err: mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"DotNotationDollarInKey": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"v.$foo", true}},
			err: mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"PrefixSuffix": {
			filter:     bson.D{{"_id", "array-numbers-asc"}},
			projection: bson.D{{"$.foo.$", true}},
			err: mongo.CommandError{
				Code:    16410,
				Name:    "Location16410",
				Message: "FieldPath field names may not start with '$'. Consider using $getField or $setField.",
			},
		},
		"Exclusion": {
			filter:     bson.D{{"v", 42}},
			projection: bson.D{{"v.$", false}},
			err: mongo.CommandError{
				Code:    31395,
				Name:    "Location31395",
				Message: "positional projection cannot be used with exclusion",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			require.NotNil(t, tc.filter, "filter should be set")
			require.NotNil(t, tc.projection, "projection should be set")

			_, err := coll.Find(ctx, tc.filter, options.Find().SetProjection(tc.projection))
			AssertEqualAltCommandError(t, tc.err, tc.altMessage, err)
		})
	}
}
