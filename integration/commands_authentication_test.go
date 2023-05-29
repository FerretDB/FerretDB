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

	"github.com/FerretDB/FerretDB/integration/setup"
)

type connectionStatus struct {
	AuthInfo struct {
		AuthenticatedUsers []any
	}
	Ok float64
}

func TestCommandsAuthenticationLogout(t *testing.T) {
	t.Parallel()
	setup.SkipForMongoDB(t, "tls is not enabled for mongodb backend")
	setup.SkipForTigrisWithReason(t, "tls is not enabled for tigris backend")
	ctx, collection := setup.Setup(t)

	db := collection.Database()

	var statusSignedIn connectionStatus
	err := db.RunCommand(ctx, bson.D{{"connectionStatus", 1}}).Decode(&statusSignedIn)
	assert.NoError(t, err)
	assert.NotEmpty(t, statusSignedIn.AuthInfo.AuthenticatedUsers)

	var res bson.D
	err = db.RunCommand(ctx, bson.D{{"logout", 1}}).Decode(&res)

	assert.NoError(t, err)
	assert.Equal(t, bson.D{{"ok", 1.0}}, res)

	var statusSignedOut connectionStatus
	err = db.RunCommand(ctx, bson.D{{"connectionStatus", 1}}).Decode(&statusSignedOut)
	assert.NoError(t, err)
	assert.Empty(t, statusSignedOut.AuthInfo.AuthenticatedUsers)
}

func TestCommandsAuthenticationSASLStart(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB/issues/1568
}
