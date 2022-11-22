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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// TestInsertSimple checks simple cases of doc deletion.
func TestInsertSimple(t *testing.T) {
	// TODO: remove this test when validation is done.

	t.Parallel()
	ctx, collection := setup.Setup(t)

	_, err := collection.InsertOne(ctx, bson.D{{"v", math.Round(-1 / 1000000000000000000.0)}})
	require.NoError(t, err)

	res := collection.FindOne(ctx, bson.D{{"v", math.Round(-1 / 1000000000000000000.0)}})
	require.NoError(t, res.Err())

	var doc bson.D
	err = res.Decode(&doc)
	require.NoError(t, err)

	assert.Equal(t, -0.0, doc.Map()["v"])
}
