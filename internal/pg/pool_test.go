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

// Use _test package to avoid import cycle with testutil.
package pg_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestValidUTF8Locale(t *testing.T) {
	t.Parallel()

	cases := []struct {
		locale   string
		expected bool
	}{
		{"en_US.utf8", true},
		{"en_US.utf-8", true},
		{"en_US.UTF8", true},
		{"en_US.UTF-8", true},
		{"en_UK.UTF-8", false},
		{"en_UK.utf--8", false},
		{"en_US", false},
		{"utf8", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.locale, func(t *testing.T) {
			t.Parallel()

			actual := pg.IsValidUTF8Locale(tc.locale)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestTables(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t, nil, zaptest.NewLogger(t))

	tables, storages, err := pool.Tables(ctx, "monila")
	require.NoError(t, err)

	expectedTables := []string{
		"actor",
		"address",
		"category",
		"city",
		"country",
		"customer",
		"film",
		"film_actor",
		"film_category",
		"inventory",
		"language",
		"rental",
		"staff",
		"store",
	}
	assert.Equal(t, expectedTables, tables)
	assert.Len(t, storages, len(expectedTables))
	for _, s := range storages {
		assert.Equal(t, pg.JSONB1Table, s)
	}

	tables, storages, err = pool.Tables(ctx, "pagila")
	require.NoError(t, err)

	expectedTables = []string{
		"actor",
		"actor_info",
		"address",
		"category",
		"city",
		"country",
		"customer",
		"customer_list",
		"film",
		"film_actor",
		"film_category",
		"film_list",
		"inventory",
		"language",
		"nicer_but_slower_film_list",
		"payment",
		"payment_p2020_01",
		"payment_p2020_02",
		"payment_p2020_03",
		"payment_p2020_04",
		"payment_p2020_05",
		"payment_p2020_06",
		"rental",
		"sales_by_film_category",
		"sales_by_store",
		"staff",
		"staff_list",
		"store",
	}
	assert.Equal(t, expectedTables, tables)
	assert.Len(t, storages, len(expectedTables))
	for _, s := range storages {
		assert.Equal(t, pg.SQLTable, s)
	}
}

func TestConcurrentCreate(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	createPool := testutil.Pool(ctx, t, nil, zaptest.NewLogger(t))
	dbName := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_")
	_, err := createPool.Exec(ctx, `CREATE DATABASE `+dbName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := createPool.Exec(ctx, `DROP DATABASE `+dbName)
		require.NoError(t, err)
	})

	n := 10
	dsn := fmt.Sprintf("postgres://postgres@127.0.0.1:5432/%[1]s?pool_min_conns=%[2]d&pool_max_conns=%[2]d", dbName, n)
	pool, err := pg.NewPool(dsn, zaptest.NewLogger(t), false)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	schemaName := testutil.SchemaName(t)
	tableName := testutil.TableName(t)

	for _, withTable := range []bool{false, true} {
		start := make(chan struct{})
		res := make(chan error, n)
		for i := 0; i < n; i++ {
			go func() {
				<-start
				if withTable {
					res <- pool.CreateTable(ctx, schemaName, tableName)
				} else {
					res <- pool.CreateSchema(ctx, schemaName)
				}
			}()
		}

		close(start)

		var errors int
		for i := 0; i < n; i++ {
			err := <-res
			if err == nil {
				continue
			}

			errors++
			assert.Equal(t, pg.ErrAlreadyExist, err)
		}

		assert.Equal(t, n-1, errors)

		// one more time to check "normal" error (DuplicateSchema, DuplicateTable)
		if withTable {
			assert.Equal(t, pg.ErrAlreadyExist, pool.CreateTable(ctx, schemaName, tableName))
		} else {
			assert.Equal(t, pg.ErrAlreadyExist, pool.CreateSchema(ctx, schemaName))
		}
	}
}
