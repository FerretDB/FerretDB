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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestUpdateArrayPush(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Composites)

	filter := bson.D{{"_id", "document-composite"}}

	update := bson.D{{"$push", bson.D{{"value.foo", "baz"}}}}
	res, err := collection.UpdateOne(ctx, filter, update)
	expectedErr := mongo.WriteException{
		WriteErrors: mongo.WriteErrors{{
			Code:    2,
			Message: `The field 'value.foo' must be an array but is of type int in document {_id: "document-composite"}`,
		}},
	}
	AssertEqualError(t, expectedErr, err)
	assert.Nil(t, res)
}
