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

package pgdb

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCreateIndexIfNotExists(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t)

	databaseName := testutil.DatabaseName(t)
	collectionName := testutil.CollectionName(t)
	setupDatabase(ctx, t, pool, databaseName)

	indexName := "test"
	err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
		idx := Index{
			Name: indexName,
			Key:  []IndexKeyPair{{Field: "foo", Order: types.Ascending}, {Field: "bar", Order: types.Descending}},
		}
		return CreateIndexIfNotExists(ctx, tx, databaseName, collectionName, &idx)
	})
	require.NoError(t, err)

	tableName := collectionNameToTableName(collectionName)
	pgIndexName := indexNameToPgIndexName(collectionName, indexName)

	var indexdef string
	err = pool.p.QueryRow(
		ctx,
		"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
		databaseName, tableName, pgIndexName,
	).Scan(&indexdef)
	require.NoError(t, err)

	expectedIndexdef := fmt.Sprintf(
		"CREATE INDEX %s ON \"%s\".%s USING btree (((_jsonb -> 'foo'::text)), ((_jsonb -> 'bar'::text)) DESC)",
		pgIndexName, databaseName, tableName,
	)
	assert.Equal(t, expectedIndexdef, indexdef)
}

// TestDropIndexes checks that we correctly drop indexes for various combination of existing indexes.
func TestDropIndexes(t *testing.T) {
	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t)

	databaseName := testutil.DatabaseName(t)
	collectionName := testutil.CollectionName(t)
	setupDatabase(ctx, t, pool, databaseName)

	err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
		return CreateCollectionIfNotExists(ctx, tx, databaseName, collectionName)
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		expectedErr error   // expected error, if any
		toCreate    []Index // indexes to create before dropping
		toDrop      []Index // indexes to drop
		expected    []Index // expected indexes to remain after dropping attempt
	}{
		"NonExistent": {
			toCreate: []Index{},
			toDrop:   []Index{{Name: "foo_1"}},
			expected: []Index{
				{Name: "_id_", Key: []IndexKeyPair{{Field: "_id", Order: types.Ascending}}, Unique: true},
			},
			expectedErr: ErrIndexNotExist,
		},
		"DropOneByName": {
			toCreate: []Index{
				{Name: "foo_1", Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}},
			},
			toDrop: []Index{{Name: "foo_1"}},
			expected: []Index{
				{Name: "_id_", Key: []IndexKeyPair{{Field: "_id", Order: types.Ascending}}, Unique: true},
			},
		},
		"DropOneByKey": {
			toCreate: []Index{
				{Name: "foo_1", Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}},
			},
			toDrop: []Index{{Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}}},
			expected: []Index{
				{Name: "_id_", Key: []IndexKeyPair{{Field: "_id", Order: types.Ascending}}, Unique: true},
			},
		},
		"DropOneFromTheBeginning": {
			toCreate: []Index{
				{Name: "foo_1", Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}},
				{Name: "bar_1", Key: []IndexKeyPair{{Field: "bar", Order: types.Ascending}}},
				{Name: "car_1", Key: []IndexKeyPair{{Field: "car", Order: types.Ascending}}},
			},
			toDrop: []Index{{Name: "foo_1"}},
			expected: []Index{
				{Name: "_id_", Key: []IndexKeyPair{{Field: "_id", Order: types.Ascending}}, Unique: true},
				{Name: "bar_1", Key: []IndexKeyPair{{Field: "bar", Order: types.Ascending}}},
				{Name: "car_1", Key: []IndexKeyPair{{Field: "car", Order: types.Ascending}}},
			},
		},
		"DropOneFromTheMiddle": {
			toCreate: []Index{
				{Name: "foo_1", Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}},
				{Name: "bar_1", Key: []IndexKeyPair{{Field: "bar", Order: types.Ascending}}},
				{Name: "car_1", Key: []IndexKeyPair{{Field: "car", Order: types.Ascending}}},
			},
			toDrop: []Index{{Name: "bar_1"}},
			expected: []Index{
				{Name: "_id_", Key: []IndexKeyPair{{Field: "_id", Order: types.Ascending}}, Unique: true},
				{Name: "foo_1", Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}},
				{Name: "car_1", Key: []IndexKeyPair{{Field: "car", Order: types.Ascending}}},
			},
		},
		"DropOneFromTheEnd": {
			toCreate: []Index{
				{Name: "foo_1", Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}},
				{Name: "bar_1", Key: []IndexKeyPair{{Field: "bar", Order: types.Ascending}}},
				{Name: "car_1", Key: []IndexKeyPair{{Field: "car", Order: types.Ascending}}},
			},
			toDrop: []Index{{Name: "car_1"}},
			expected: []Index{
				{Name: "_id_", Key: []IndexKeyPair{{Field: "_id", Order: types.Ascending}}, Unique: true},
				{Name: "foo_1", Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}},
				{Name: "bar_1", Key: []IndexKeyPair{{Field: "bar", Order: types.Ascending}}},
			},
		},
		"DropTwo": {
			toCreate: []Index{
				{Name: "foo_1", Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}},
				{Name: "bar_1", Key: []IndexKeyPair{{Field: "bar", Order: types.Ascending}}},
				{Name: "car_1", Key: []IndexKeyPair{{Field: "car", Order: types.Ascending}}},
			},
			toDrop: []Index{{Name: "car_1"}, {Name: "foo_1"}},
			expected: []Index{
				{Name: "_id_", Key: []IndexKeyPair{{Field: "_id", Order: types.Ascending}}, Unique: true},
				{Name: "bar_1", Key: []IndexKeyPair{{Field: "bar", Order: types.Ascending}}},
			},
		},
		"DropComplicated": {
			toCreate: []Index{
				{Name: "v_-1", Key: []IndexKeyPair{{Field: "v", Order: types.Descending}}},
				{Name: "v_1_foo_1", Key: []IndexKeyPair{{Field: "v", Order: types.Ascending}, {Field: "foo", Order: types.Ascending}}}, //nolint:lll // for readability
				{Name: "v.foo_-1", Key: []IndexKeyPair{{Field: "v.foo", Order: types.Descending}}},
			},
			toDrop: []Index{{Name: "v_-1"}, {Name: "v_1_foo_1"}},
			expected: []Index{
				{Name: "_id_", Key: []IndexKeyPair{{Field: "_id", Order: types.Ascending}}, Unique: true},
				{Name: "v.foo_-1", Key: []IndexKeyPair{{Field: "v.foo", Order: types.Descending}}},
			},
		},
		"DropAll": {
			toCreate: []Index{
				{Name: "foo_1", Key: []IndexKeyPair{{Field: "foo", Order: types.Ascending}}},
				{Name: "bar_1", Key: []IndexKeyPair{{Field: "bar", Order: types.Ascending}}},
				{Name: "car_1", Key: []IndexKeyPair{{Field: "car", Order: types.Ascending}}},
			},
			toDrop: []Index{{Name: "bar_1"}, {Name: "car_1"}, {Name: "foo_1"}},
			expected: []Index{
				{Name: "_id_", Key: []IndexKeyPair{{Field: "_id", Order: types.Ascending}}, Unique: true},
			},
		},
	} {
		tc := tc

		// We don't run this subtest in parallel because we use the same database and collection.
		t.Run(name, func(t *testing.T) {
			t.Helper()

			err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
				for _, idx := range tc.toCreate {
					if err := CreateIndexIfNotExists(ctx, tx, databaseName, collectionName, &idx); err != nil {
						return err
					}
				}

				return nil
			})
			require.NoError(t, err)

			expectedWas := int32(len(tc.toCreate) + 1) // created indexes + default _id index
			err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
				for _, idx := range tc.toDrop {
					var was int32
					was, err = DropIndex(ctx, tx, databaseName, collectionName, &idx)
					if err != nil {
						return err
					}

					assert.Equal(t, expectedWas, was)
					expectedWas--
				}

				return nil
			})

			if tc.expectedErr != nil {
				assert.True(t, errors.Is(err, tc.expectedErr))
			} else {
				require.NoError(t, err)
			}

			err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
				var indexes []Index

				indexes, err = Indexes(ctx, tx, databaseName, collectionName)
				if err != nil {
					return err
				}

				assert.Equal(t, tc.expected, indexes)
				return nil
			})
			require.NoError(t, err)

			err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
				var was int32
				was, err = DropAllIndexes(ctx, tx, databaseName, collectionName)
				if err != nil {
					return err
				}

				assert.Equal(t, expectedWas, was)

				var indexes []Index
				indexes, err = Indexes(ctx, tx, databaseName, collectionName)
				if err != nil {
					return err
				}

				assert.Len(t, indexes, 1) // only default _id index left
				return nil
			})
			require.NoError(t, err)
		})
	}
}

func TestDropIndexesStress(t *testing.T) {
	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t)

	databaseName := testutil.DatabaseName(t)
	collectionName := testutil.CollectionName(t)
	setupDatabase(ctx, t, pool, databaseName)

	var initialIndexes []Index
	var err error

	err = pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		if err = CreateCollection(ctx, tx, databaseName, collectionName); err != nil {
			return err
		}

		initialIndexes, err = Indexes(ctx, tx, databaseName, collectionName)
		return err
	})
	require.NoError(t, err)

	indexName := "test"
	indexKeys := []IndexKeyPair{{Field: "foo", Order: types.Ascending}, {Field: "bar", Order: types.Descending}}

	err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
		idx := Index{
			Name: indexName,
			Key:  indexKeys,
		}

		return CreateIndexIfNotExists(ctx, tx, databaseName, collectionName, &idx)
	})
	require.NoError(t, err)

	var indexesAfterCreate []Index

	err = pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		indexesAfterCreate, err = Indexes(ctx, tx, databaseName, collectionName)
		return err
	})
	require.NoError(t, err)

	dropNum := runtime.GOMAXPROCS(-1) * 10

	ready := make(chan struct{}, dropNum)
	start := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i <= dropNum; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			ready <- struct{}{}

			<-start

			// do not use `err`, to avoid data race
			tErr := pool.InTransaction(ctx, func(tx pgx.Tx) error {
				idx := Index{
					Name: indexName,
					Key:  indexKeys,
				}

				// do not use `err`, to avoid data race
				_, dropErr := DropIndex(ctx, tx, databaseName, collectionName, &idx)
				return dropErr
			})
			// if the index could not be dropped, the error is checked
			if tErr != nil {
				require.Error(t, tErr, ErrIndexNotExist)
			}
		}()
	}

	for i := 0; i < dropNum; i++ {
		<-ready
	}

	close(start)

	wg.Wait()

	var indexesAfterDrop []Index

	err = pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		indexesAfterDrop, err = Indexes(ctx, tx, databaseName, collectionName)
		return err
	})
	require.NoError(t, err)

	require.Equal(t, initialIndexes, indexesAfterDrop)
	require.NotEqual(t, indexesAfterCreate, indexesAfterDrop)
}
