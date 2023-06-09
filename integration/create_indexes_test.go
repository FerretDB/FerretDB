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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestCreateIndexesErrors(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes creation is not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		models     []mongo.IndexModel  // required, index model to create
		err        *mongo.CommandError // required, expected error
		altMessage string              // optional, alternative error message
		skip       string              // optional, skip test with a specified reason
	}{
		"UniqueOneIndex": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"v", 1}},
					Options: new(options.IndexOptions).SetUnique(true),
				},
			},

			err: &mongo.CommandError{
				Code: 11000,
				Name: "DuplicateKey",
				Message: "Index build failed\\: " +
					"[0-9a-fA-F]{8}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{12}\\: Collection " +
					"TestCreateIndexesErrors-UniqueOneIndex\\.TestCreateIndexesErrors-UniqueOneIndex " +
					"\\( [0-9a-fA-F]{8}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{12} \\) " +
					"\\:\\: caused by \\:\\: E11000 duplicate key error collection\\: " +
					"TestCreateIndexesErrors-UniqueOneIndex\\.TestCreateIndexesErrors-UniqueOneIndex " +
					"index\\: v_1 dup key\\: \\{ v\\: null \\}",
			},
			altMessage: "Index build failed",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Helper()
			t.Parallel()

			ctx, collection := setup.Setup(t, shareddata.Composites)

			cur, err := collection.Indexes().List(ctx)
			require.NoError(t, err)

			res := FetchAll(t, ctx, cur)

			fmt.Println(res)

			_, err = collection.Indexes().CreateMany(ctx, tc.models)
			//if tc.err != nil {
			//	AssertRegexCommandError(t, *tc.err, err)
			//
			//	return
			//}

			require.NoError(t, err)
		})
	}
}
