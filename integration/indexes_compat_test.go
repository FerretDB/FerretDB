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

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestIndexesList(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                shareddata.AllProviders(),
		AddNonExistentCollection: true,
	})
	ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

	for i := range targetCollections {
		targetCollection := targetCollections[i]
		compatCollection := compatCollections[i]

		t.Run(targetCollection.Name(), func(t *testing.T) {
			t.Helper()
			t.Parallel()

			targetCur, targetErr := targetCollection.Indexes().List(ctx)
			compatCur, compatErr := compatCollection.Indexes().List(ctx)

			require.NoError(t, compatErr)
			require.NoError(t, targetErr)

			targetRes := FetchAll(t, ctx, targetCur)
			compatRes := FetchAll(t, ctx, compatCur)

			assert.Equal(t, compatRes, targetRes)

			// Also test specifications to check they are identical.
			targetSpec, targetErr := targetCollection.Indexes().ListSpecifications(ctx)
			compatSpec, compatErr := compatCollection.Indexes().ListSpecifications(ctx)

			require.NoError(t, compatErr)
			require.NoError(t, targetErr)

			assert.Equal(t, compatSpec, targetSpec)
		})
	}
}
