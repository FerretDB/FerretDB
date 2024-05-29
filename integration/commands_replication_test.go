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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/internal/handler/common"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCommandsReplication(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	for _, command := range []string{"ismaster", "isMaster", "hello"} {
		command := command
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, bson.D{{command, 1}}).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()
			t.Log(m)

			delete(m, "ismaster")
			delete(m, "connectionId")
			delete(m, "logicalSessionTimeoutMinutes")
			delete(m, "topologyVersion")
			delete(m, "isWritablePrimary")
			delete(m, "$clusterTime")
			delete(m, "electionId")
			delete(m, "hosts")
			delete(m, "lastWrite")
			delete(m, "me")
			delete(m, "operationTime")
			delete(m, "primary")
			delete(m, "secondary")
			delete(m, "setName")
			delete(m, "setVersion")

			assert.InDelta(t, time.Now().Unix(), m["localTime"].(primitive.DateTime).Time().Unix(), 2)
			delete(m, "localTime")

			maxWireVersion := m["maxWireVersion"].(int32)
			assert.Equal(t, common.MaxWireVersion, maxWireVersion)
			delete(m, "maxWireVersion")

			minWireVersion := m["minWireVersion"].(int32)
			assert.True(t, minWireVersion == 0 || minWireVersion == common.MinWireVersion)
			delete(m, "minWireVersion")

			expected := bson.M{
				"maxBsonObjectSize":   int32(16777216),
				"maxMessageSizeBytes": int32(48000000),
				"maxWriteBatchSize":   int32(100000),
				"ok":                  float64(1),
				"readOnly":            false,
			}

			assert.Equal(t, expected, m)
		})
	}
}
