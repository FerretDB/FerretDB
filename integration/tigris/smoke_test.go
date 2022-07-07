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

package tigris

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestSmoke(t *testing.T) {
	t.Parallel()

	for i, p := range []shareddata.Provider{
		// shareddata.FixedScalars,
		shareddata.FixedScalarsIDs,
	} {
		i, p := i, p
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			ctx, collection := integration.Setup(t, p)

			doc := p.Docs()[0]
			id := doc.Map()["_id"]
			var actualDoc bson.D
			err := collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&actualDoc)
			require.NoError(t, err)
			integration.AssertEqualDocuments(t, doc, actualDoc)

			updateRes, err := collection.UpdateByID(ctx, id, bson.D{{"$set", bson.D{{"double_value", 43.13}}}})
			require.NoError(t, err)
			expectedUpdateRes := &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
			}
			assert.Equal(t, expectedUpdateRes, updateRes)

			deleteRes, err := collection.DeleteOne(ctx, bson.D{{"_id", id}})
			require.NoError(t, err)
			expectedDeleteRes := &mongo.DeleteResult{
				DeletedCount: 1,
			}
			assert.Equal(t, expectedDeleteRes, deleteRes)
		})
	}
}
