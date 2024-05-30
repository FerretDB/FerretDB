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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestDiffUpdateProduceInfinity(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)
	_, err := collection.InsertOne(ctx, bson.D{{"_id", "number"}, {"v", int32(42)}})
	require.NoError(t, err)

	_, err = collection.UpdateOne(ctx, bson.D{{"_id", "number"}}, bson.D{{"$mul", bson.D{{"v", math.MaxFloat64}}}})

	if setup.IsMongoDB(t) {
		require.NoError(t, err)
		return
	}

	expected := mongo.CommandError{
		Code: 2,
		Name: "BadValue",
		Message: `update produces invalid value: { "v": +Inf }` +
			` (update operations that produce infinity values are not allowed)`,
	}
	AssertEqualCommandError(t, expected, err)
}
