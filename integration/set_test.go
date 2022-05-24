// Copyright 2021 FerretDB Set.
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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

type testCase struct {
	filter bson.D
	update bson.D
	err    *mongo.WriteError
	alt    string
}

func TestSetOperator(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]testCase{
		"BadSetString": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$set", "string"}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: \"string\"}",
			},
		},
		"BadSetDouble": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$set", float64(42.12345)}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: 42.12345}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: 42.12}",
		},
		"BadSetArray": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$set", bson.A{}}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: []}",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := collection.UpdateOne(ctx, tc.filter, tc.update)
			if tc.err != nil {
				t.Log(err)
				AssertEqualWriteError(t, tc.err, tc.alt, err)
				return
			}
		})
	}
}
