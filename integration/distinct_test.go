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

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestDistinctErrors(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command  any                // required
		collName any                // optional
		filter   bson.D             // required
		err      mongo.CommandError // required
	}{
		"EmptyCollection": {
			command:  "a",
			filter:   bson.D{},
			collName: "",
			err: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Invalid namespace specified 'TestDistinctErrors.'",
			},
		},
		"CollectionTypeObject": {
			command:  "a",
			filter:   bson.D{},
			collName: bson.D{},
			err: mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type object",
			},
		},
		"WrongTypeObject": {
			command: bson.D{},
			filter:  bson.D{},
			err: mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.key' is the wrong type 'object', expected type 'string'",
			},
		},
		"WrongTypeArray": {
			command: bson.A{},
			filter:  bson.D{},
			err: mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.key' is the wrong type 'array', expected type 'string'",
			},
		},
		"WrongTypeNumber": {
			command: int32(1),
			filter:  bson.D{},
			err: mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.key' is the wrong type 'int', expected type 'string'",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			require.NotNil(t, tc.filter, "filter should be set")

			var collName any = coll.Name()
			if tc.collName != nil {
				collName = tc.collName
			}

			command := bson.D{{"distinct", collName}, {"key", tc.command}, {"query", tc.filter}}

			res := coll.Database().RunCommand(ctx, command)
			require.Error(t, res.Err(), "expected error")

			AssertEqualCommandError(t, tc.err, res.Err())
		})
	}
}
