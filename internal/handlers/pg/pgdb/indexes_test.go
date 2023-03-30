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
	"runtime"
	"sync"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

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

	// TODO create index
	//err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
	//	idx := Index{
	//		Name: indexName,
	//		Key:  indexKeys,
	//	}
	//
	//	return CreateIndexIfNotExists(ctx, tx, databaseName, collectionName, &idx)
	//})
	//require.NoError(t, err)

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

			err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
				idx := Index{
					Name: indexName,
					Key:  indexKeys,
				}

				return DropIndex(ctx, tx, databaseName, collectionName, &idx)
			})

			// one of drop will succeed.
			if err != nil {
				require.Error(t, err, "")
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
