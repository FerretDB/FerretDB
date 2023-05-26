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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestAggregateProjectErrors(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		err        *mongo.CommandError
		altMessage string
		skip       string
		pipeline   bson.A
	}{
		"EmptyPipeline": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{}}},
			},
			err: &mongo.CommandError{
				Code:    51272,
				Name:    "Location51272",
				Message: "Invalid $project :: caused by :: projection specification must have at least one field",
			},
		},
		"EmptyProjectionField": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{{"v", bson.D{}}}}},
			},
			err: &mongo.CommandError{
				Code:    51270,
				Name:    "Location51270",
				Message: "Invalid $project :: caused by :: An empty sub-projection is not a valid value. Found empty object at path",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2633",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()
			ctx, collection := setup.Setup(t)

			_, err := collection.Aggregate(ctx, tc.pipeline)

			if tc.err != nil {
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
