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

// Package metadata provides access to SQLite databases and collections information.
package metadata

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"

	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata/pool"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// metadataTableName is a SQLite table name where FerretDB metadata is stored.
const metadataTableName = "_ferretdb_collections"

// Registry provides access to SQLite databases and collections information.
type Registry struct {
	p *pool.Pool
	l *zap.Logger
}

// NewRegistry creates a registry for the given directory.
func NewRegistry(u string, l *zap.Logger) (*Registry, error) {
	p, err := pool.New(u, l.Named("pool"))
	if err != nil {
		return nil, err
	}

	return &Registry{
		p: p,
		l: l,
	}, nil
}

// Close closes the registry.
func (r *Registry) Close() {
	r.p.Close()
}

// CollectionToTable converts FerretDB collection name to SQLite table name.
func (r *Registry) CollectionToTable(collectionName string) string {
	// TODO https://github.com/FerretDB/FerretDB/issues/2749
	h := sha1.Sum([]byte(collectionName))
	return hex.EncodeToString(h[:])
}

// DatabaseList returns a sorted list of existing databases.
func (r *Registry) DatabaseList(ctx context.Context) []string {
	return r.p.List(ctx)
}

// DatabaseGetExisting returns a connection to existing database or nil if it doesn't exist.
func (r *Registry) DatabaseGetExisting(ctx context.Context, dbName string) *sql.DB {
	return r.p.GetExisting(ctx, dbName)
}

// DatabaseGetOrCreate returns a connection to existing database or newly created database.
func (r *Registry) DatabaseGetOrCreate(ctx context.Context, dbName string) (*sql.DB, error) {
	db, created, err := r.p.GetOrCreate(ctx, dbName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !created {
		return db, nil
	}

	// TODO create unique indexes for name and table_name https://github.com/FerretDB/FerretDB/issues/2747
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE %q (name, table_name, settings TEXT)", metadataTableName))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return db, nil
}

// DatabaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
func (r *Registry) DatabaseDrop(ctx context.Context, dbName string) bool {
	return r.p.Drop(ctx, dbName)
}

// CollectionList returns a sorted list of collections in the database.
//
// If database does not exist, no error is returned.
func (r *Registry) CollectionList(ctx context.Context, dbName string) ([]string, error) {
	db := r.p.GetExisting(ctx, dbName)
	if db == nil {
		return nil, nil
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT name FROM %q ORDER BY name", metadataTableName))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	var res []string

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, lazyerrors.Error(err)
		}

		res = append(res, name)
	}

	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// CollectionCreate creates a collection in the database.
//
// Returned boolean value indicates whether the collection was created.
// If collection already exists, (false, nil) is returned.
func (r *Registry) CollectionCreate(ctx context.Context, dbName string, collectionName string) (bool, error) {
	db, err := r.DatabaseGetOrCreate(ctx, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	tableName := r.CollectionToTable(collectionName)

	err = inTransaction(ctx, db, func(tx *sql.Tx) error {
		query := fmt.Sprintf("CREATE TABLE %q (_ferretdb_sjson TEXT)", tableName)
		if _, err = tx.ExecContext(ctx, query); err != nil {
			var e *sqlite.Error
			if errors.As(err, &e) && e.Code() == sqlitelib.SQLITE_ERROR {
				return nil
			}

			return lazyerrors.Error(err)
		}

		query = fmt.Sprintf("INSERT INTO %q (name, table_name) VALUES (?, ?)", metadataTableName)
		if _, err = tx.ExecContext(ctx, query, collectionName, tableName); err != nil {
			return lazyerrors.Error(err)
		}

		return nil
	})
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	return true, nil
}

// CollectionDrop drops a collection in the database.
//
// Returned boolean value indicates whether the collection was dropped.
// If database or collection did not exist, (false, nil) is returned.
func (r *Registry) CollectionDrop(ctx context.Context, dbName string, collectionName string) (bool, error) {
	db := r.p.GetExisting(ctx, dbName)
	if db == nil {
		return false, nil
	}

	tableName := r.CollectionToTable(collectionName)

	err := inTransaction(ctx, db, func(tx *sql.Tx) error {
		query := fmt.Sprintf("DELETE FROM %q WHERE name = ?", metadataTableName)
		if _, err := tx.ExecContext(ctx, query, collectionName); err != nil {
			return lazyerrors.Error(err)
		}

		query = fmt.Sprintf("DROP TABLE %q", tableName)
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return lazyerrors.Error(err)
		}

		return nil
	})
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	return true, nil
}

// inTransaction wraps the given function f in a SQLite transaction.
//
// If f returns an error, the transaction is rolled back.
//
// Passed context will be used for transaction.
// Context cancellation DOES rollback the transaction.
func inTransaction(ctx context.Context, db *sql.DB, f func(*sql.Tx) error) (err error) {
	var tx *sql.Tx

	if tx, err = db.BeginTx(ctx, nil); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	if err = f(tx); err != nil {
		err = lazyerrors.Error(err)
		_ = tx.Rollback()

		return
	}

	if err = tx.Commit(); err != nil {
		err = lazyerrors.Error(err)
		_ = tx.Rollback()

		return
	}

	return
}
