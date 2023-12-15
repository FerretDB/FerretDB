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

package oplog

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestOplogInsert(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t)

	_, err := coll.InsertOne(ctx, bson.D{{"foo", "bar"}})
	require.NoError(t, err)

	local := coll.Database().Client().Database("local")

	var lastOplogEntry bson.D

	opts := options.FindOne().SetSort(bson.D{{"$natural", -1}})
	err = local.Collection("oplog.rs").FindOne(ctx, bson.D{}, opts).Decode(&lastOplogEntry)
	require.NoError(t, err)

	fmt.Println(lastOplogEntry)
}
