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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCommandsFreeMonitoringGetFreeMonitoringStatus(t *testing.T) {
	t.Parallel()
	ctx, collection := setupWithOpts(t, &setupOpts{
		databaseName: "admin",
	})

	expected := map[string]any{
		"state": "disabled",
		"ok":    float64(1),
	}

	var actual bson.D
	err := collection.Database().RunCommand(ctx, bson.D{{"getFreeMonitoringStatus", 1}}).Decode(&actual)
	require.NoError(t, err)

	m := actual.Map()
	keys := CollectKeys(t, actual)

	for k, item := range expected {
		assert.Contains(t, keys, k)
		assert.IsType(t, item, m[k])

		if it, ok := item.(primitive.D); ok {
			z := m[k].(primitive.D)
			AssertEqualDocuments(t, it, z)
			continue
		}
		assert.Equal(t, m[k], item)
	}
}
