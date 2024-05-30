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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

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
