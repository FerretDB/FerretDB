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
	"context"
	"fmt"
	"hash/fnv"

	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// Internal collections prefix.
	collectionPrefix = "_ferretdb_"

	// PostgreSQL max table name length.
	maxTableNameLength = 63
)

// CreateSettingsTable creates FerretDB settings table.
func (pgPool *Pool) CreateSettingsTable(ctx context.Context, db string) error {
	tables, err := pgPool.tables(ctx, db)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if slices.Contains(tables, collectionPrefix+"settings") {
		return ErrAlreadyExist
	}

	sql := `CREATE TABLE ` + pgx.Identifier{db, collectionPrefix + "settings"}.Sanitize() + ` (settings jsonb)`
	_, err = pgPool.Exec(ctx, sql)
	if err != nil {
		return err
	}

	settings := must.NotFail(types.NewDocument("collections", must.NotFail(types.NewDocument())))
	sql = fmt.Sprintf(`INSERT INTO %s (settings) VALUES ($1)`, pgx.Identifier{db, collectionPrefix + "settings"}.Sanitize())
	_, err = pgPool.Exec(ctx, sql, must.NotFail(fjson.Marshal(settings)))
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// GetTableName returns the name of the table for given collection or error.
func (pgPool *Pool) GetTableName(ctx context.Context, db, collection string) (string, error) {
	if err := pgPool.CreateSchema(ctx, db); err != nil && err != ErrAlreadyExist {
		return "", lazyerrors.Error(err)
	}

	var err error
	var tables []string
	if tables, err = pgPool.tables(ctx, db); err != nil {
		return "", lazyerrors.Error(err)
	}
	if !slices.Contains(tables, collectionPrefix+"settings") {
		err = pgPool.CreateSettingsTable(ctx, db)
		if err != nil {
			return "", lazyerrors.Error(err)
		}
	}

	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return "", lazyerrors.Error(err)
	}
	defer func() {
		if err != nil {
			must.NoError(tx.Rollback(ctx))
			return
		}
		must.NoError(tx.Commit(ctx))
	}()

	settings, err := pgPool.getSettingsTable(ctx, tx, db)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	collections, ok := must.NotFail(settings.Get("collections")).(*types.Document)
	if !ok {
		return "", fmt.Errorf("invalid settings document")
	}

	if collections.Has(collection) {
		return must.NotFail(collections.Get(collection)).(string), nil
	}

	tableName := GetTableNameFormatted(collection)
	must.NoError(collections.Set(collection, tableName))
	must.NoError(settings.Set("collections", collections))

	err = pgPool.updateSettingsTable(ctx, tx, db, settings)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	return tableName, nil
}

func (pgPool *Pool) updateSettingsTable(ctx context.Context, tx pgx.Tx, db string, settings *types.Document) error {
	sql := `UPDATE ` + pgx.Identifier{db, collectionPrefix + "settings"}.Sanitize() + `SET settings = $1`
	_, err := tx.Exec(ctx, sql, must.NotFail(fjson.Marshal(settings)))
	return err
}

func (pgPool *Pool) getSettingsTable(ctx context.Context, tx pgx.Tx, db string) (*types.Document, error) {
	sql := `SELECT settings FROM ` + pgx.Identifier{db, collectionPrefix + "settings"}.Sanitize()
	rows, err := tx.Query(ctx, sql)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no rows returned")
	}

	var b []byte
	if err := rows.Scan(&b); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := fjson.Unmarshal(b)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	settings, ok := doc.(*types.Document)
	if !ok {
		return nil, fmt.Errorf("invalid settings document")
	}

	return settings, nil
}

func (pgPool *Pool) RemoveTableFromSettings(ctx context.Context, db, collection string) error {
	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer func() {
		if err != nil {
			must.NoError(tx.Rollback(ctx))
			return
		}
		must.NoError(tx.Commit(ctx))
	}()

	settings, err := pgPool.getSettingsTable(ctx, tx, db)
	if err != nil {
		return lazyerrors.Error(err)
	}

	collections, ok := must.NotFail(settings.Get("collections")).(*types.Document)
	if !ok {
		return fmt.Errorf("invalid settings document")
	}

	collections.Remove(collection)

	must.NoError(settings.Set("collections", collections))

	if err := pgPool.updateSettingsTable(ctx, tx, db, settings); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// GetTableNameFormatted returns collection name in form <shortened_name>_<name_hash>.
func GetTableNameFormatted(name string) string {
	hash32 := fnv.New32a()
	_ = must.NotFail(hash32.Write([]byte(name)))

	nameSymbolsLeft := maxTableNameLength - hash32.Size()*2 - 1
	truncateTo := len(name)
	if truncateTo > nameSymbolsLeft {
		truncateTo = nameSymbolsLeft
	}

	return name[:truncateTo] + "_" + fmt.Sprintf("%x", hash32.Sum([]byte{}))
}
