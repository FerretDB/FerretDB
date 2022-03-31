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

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryComparisionEq(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars)

	for _, expected := range shareddata.Scalars.Docs() {
		id := expected.Map()["_id"]
		var actual bson.D
		err := collection.FindOne(ctx, bson.D{{"_id", bson.D{{"$eq", id}}}}).Decode(&actual)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	}
}

// $gt

// $gte

// $in

// $lt

// $lte

// $ne

// $nin
