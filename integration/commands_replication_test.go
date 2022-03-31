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
)

func TestCommandsReplicationIsMaster(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	for _, command := range []string{"ismaster", "isMaster"} {
		command := command
		t.Run(command, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, bson.D{{command, 1}}).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()
			t.Log(m)
			assert.Equal(t, true, m["ismaster"])
		})
	}
}
