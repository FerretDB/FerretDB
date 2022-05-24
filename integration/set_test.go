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

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// AssertEqualWriteError compares expected mongo.WriteError Message and Code with actual error.
func AssertEqualWriteError(t *testing.T, expected *mongo.WriteError, actual error) bool {
	t.Helper()

	actualEx, ok := actual.(mongo.WriteException)
	if ok {
		if len(actualEx.WriteErrors) != 1 {
			return assert.Equal(t, expected, actual)
		}
		actualWE := actualEx.WriteErrors[0]
		return assert.Equal(t, expected.Message, actualWE.Message) &&
			assert.Equal(t, int(expected.Code), int(actualWE.Code))
	}

	actualE, ok := actual.(mongo.WriteError)
	if !ok {
		return assert.Equal(t, expected, actual)
	}

	return assert.Equal(t, expected.Message, actualE.Message) &&
		assert.Equal(t, expected.Code, actualE.Code)
}

func TestSetOperator(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		filter bson.D
		update bson.D
		result bson.D
		err    *mongo.WriteError
	}{
		"BadSetType": {
			filter: bson.D{{"_id", "string"}},
			update: bson.D{{"$set", "string"}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: \"string\"}",
			},
		},
		"SetArray": {
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
				AssertEqualWriteError(t, tc.err, err)
				return
			}
		})
	}
}
