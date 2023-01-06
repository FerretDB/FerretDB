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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryEvaluationRegex(t *testing.T) {
	// TODO: move to compat https://github.com/FerretDB/FerretDB/issues/1576

	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "multiline-string"}, {"v", "bar\nfoo"}},
		bson.D{
			{"_id", "document-nested-strings"},
			{"v", bson.D{{"foo", bson.D{{"bar", "quz"}}}}},
		},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      any
		expectedIDs []any
		err         *mongo.CommandError
		altMessage  string
	}{
		"Regex": {
			filter:      bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "foo"}}}}},
			expectedIDs: []any{"multiline-string", "string"},
		},
		"RegexNested": {
			filter:      bson.D{{"v.foo.bar", bson.D{{"$regex", primitive.Regex{Pattern: "quz"}}}}},
			expectedIDs: []any{"document-nested-strings"},
		},
		"RegexWithOption": {
			filter:      bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "42", Options: "i"}}}}},
			expectedIDs: []any{"string-double", "string-whole"},
		},
		"RegexStringOptionMatchCaseInsensitive": {
			filter:      bson.D{{"v", bson.D{{"$regex", "foo"}, {"$options", "i"}}}},
			expectedIDs: []any{"multiline-string", "regex", "string"},
		},
		"RegexStringOptionMatchLineEnd": {
			filter:      bson.D{{"v", bson.D{{"$regex", "b.*foo"}, {"$options", "s"}}}},
			expectedIDs: []any{"multiline-string"},
		},
		"RegexStringOptionMatchMultiline": {
			filter:      bson.D{{"v", bson.D{{"$regex", "^foo"}, {"$options", "m"}}}},
			expectedIDs: []any{"multiline-string", "string"},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}
