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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestTTLIndexRemovesExpiredDocs(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expireAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(1),
	})
	require.NoError(t, err)

	expired := bson.D{{Key: "_id", Value: "expired"}, {Key: "expireAt", Value: time.Now().Add(-5 * time.Minute)}}
	fresh := bson.D{{Key: "_id", Value: "fresh"}, {Key: "expireAt", Value: time.Now().Add(10 * time.Minute)}}

	_, err = collection.InsertMany(ctx, []any{expired, fresh})
	require.NoError(t, err)

	count, err := collection.CountDocuments(ctx, bson.D{})
	require.NoError(t, err)
	require.Equal(t, int64(2), count)

	deadline := time.Now().Add(90 * time.Second)

	for time.Now().Before(deadline) && ctx.Err() == nil {
		err = collection.FindOne(ctx, bson.D{{Key: "_id", Value: "expired"}}).Err()
		if errors.Is(err, mongo.ErrNoDocuments) {
			break
		}

		require.NoError(t, err)
		time.Sleep(2 * time.Second)
	}

	err = collection.FindOne(ctx, bson.D{{Key: "_id", Value: "expired"}}).Err()
	require.Error(t, err)
	require.ErrorIs(t, err, mongo.ErrNoDocuments)

	err = collection.FindOne(ctx, bson.D{{Key: "_id", Value: "fresh"}}).Err()
	require.NoError(t, err)
}
