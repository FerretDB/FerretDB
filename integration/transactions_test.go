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
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

// TestTransactionAbort checks that writes of an aborted transaction are not kept.
//
// MongoDB rolls them back; FerretDB does not implement transactions yet
// and must refuse the writes instead of executing them outside of the transaction.
func TestTransactionAbort(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	sess, err := collection.Database().Client().StartSession()
	require.NoError(t, err)

	defer sess.EndSession(ctx)

	require.NoError(t, sess.StartTransaction())

	_, err = collection.InsertOne(mongo.NewSessionContext(ctx, sess), bson.D{{"_id", "aborted"}})

	if setup.IsMongoDB(t) {
		require.NoError(t, err)
	} else {
		expected := mongo.CommandError{
			Code:    238,
			Name:    "NotImplemented",
			Message: "Transactions are not implemented yet",
		}
		AssertEqualCommandError(t, expected, err)
	}

	// abortTransaction errors are ignored by drivers, so the application would not notice them
	_ = sess.AbortTransaction(ctx)

	n, err := collection.CountDocuments(ctx, bson.D{})
	require.NoError(t, err)
	assert.Zero(t, n, "a write of an aborted transaction was kept")
}

// TestTransactionCommitFailed checks that writes of a transaction that failed to commit are not kept.
func TestTransactionCommitFailed(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	sess, err := collection.Database().Client().StartSession()
	require.NoError(t, err)

	defer sess.EndSession(ctx)

	_, err = sess.WithTransaction(ctx, func(sc mongo.SessionContext) (any, error) {
		return collection.InsertOne(sc, bson.D{{"_id", "committed"}})
	})

	n, countErr := collection.CountDocuments(ctx, bson.D{})
	require.NoError(t, countErr)

	if setup.IsMongoDB(t) {
		require.NoError(t, err)
		assert.EqualValues(t, 1, n)

		return
	}

	expected := mongo.CommandError{
		Code:    238,
		Name:    "NotImplemented",
		Message: "Transactions are not implemented yet",
	}
	AssertEqualCommandError(t, expected, err)

	assert.Zero(t, n, "a write of a failed transaction was kept")
}
