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
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
)

func TestAggregateProjectErrors(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		expectedErr *mongo.CommandError
		altMessage  string
		pipeline    bson.A
	}{
		"EmptyPipeline": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{}}},
			},
			expectedErr: &mongo.CommandError{
				Code:    int32(commonerrors.ErrEmptyProject),
				Name:    commonerrors.ErrEmptyProject.String(),
				Message: "Invalid $project :: caused by :: projection specification must have at least one field",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t)

			_, err := collection.Aggregate(ctx, tc.pipeline)

			if tc.expectedErr != nil {
				AssertEqualAltCommandError(t, *tc.expectedErr, tc.altMessage, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
