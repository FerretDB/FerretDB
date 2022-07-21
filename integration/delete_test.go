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

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// TestDeleteSimple checks simple cases of doc deletion.
func TestDeleteSimple(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		collection    string
		expectedCount int64
	}{
		"DeleteOne": {
			collection:    collection.Name(),
			expectedCount: 1,
		},
		"DeleteFromNonExistingCollection": {
			collection:    "doesnotexist",
			expectedCount: 0,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Database().Collection(tc.collection).DeleteOne(ctx, bson.D{})
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCount, cursor.DeletedCount)
		})
	}
}
