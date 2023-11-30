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

package backends_test // to avoid import cycle

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestListCollections(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbName := testutil.DatabaseName(t)

			db, err := b.Database(dbName)
			require.NoError(t, err)

			var res *backends.ListCollectionsResult

			t.Run("EmptyList", func(t *testing.T) {
				res, err = db.ListCollections(ctx, nil)
				require.NoError(t, err)

				require.NotNil(t, res)
				require.Empty(t, res.Collections)
			})

			t.Run("ValidUUID", func(t *testing.T) {
				collName := testutil.CollectionName(t)

				err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
					Name: collName,
				})
				require.NoError(t, err)

				res, err = db.ListCollections(ctx, nil)
				require.NoError(t, err)

				require.NotNil(t, res)
				require.NotEmpty(t, res.Collections)

				for _, coll := range res.Collections {
					require.NotEmpty(t, coll.Name)

					_, err = uuid.Parse(coll.UUID)
					require.NoError(t, err)
				}
			})
		})
	}
}
