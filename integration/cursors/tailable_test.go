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

package cursors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestCursorsTailable(t *testing.T) {
	t.Parallel()

	t.Run("NonCapped", func(t *testing.T) {
		t.Parallel()

		ctx, collection := setup.Setup(t, shareddata.Scalars)

		for _, ct := range []options.CursorType{options.Tailable, options.TailableAwait} {
			cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetCursorType(ct))
			expected := mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: "error processing query: " +
					"ns=TestTailable-NonCapped.TestTailable-NonCappedTree: $and\nSort: {}\nProj: {}\n " +
					"tailable cursor requested on non capped collection",
			}
			integration.AssertEqualAltCommandError(t, expected, "tailable cursor requested on non capped collection", err)
			assert.Nil(t, cursor)
		}
	})
}
