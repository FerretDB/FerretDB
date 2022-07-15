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

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestEnvData(t *testing.T) {
	t.Parallel()

	// see `env-data` Taskfile target
	ctx, collection, _ := SetupWithOpts(t, &SetupOpts{
		DatabaseName: "test",
	})
	collection = collection.Database().Collection("values")

	err := collection.Drop(ctx)
	require.NoError(t, err)

	providers := []shareddata.Provider{shareddata.Scalars, shareddata.Composites}
	for _, provider := range providers {
		docs := provider.Docs()
		res, err := collection.InsertMany(ctx, shareddata.DocsAny(docs))
		require.NoError(t, err)
		require.Len(t, res.InsertedIDs, len(docs))
	}
}
