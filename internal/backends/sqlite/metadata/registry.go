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
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	"go.uber.org/zap"
	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"

	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata/pool"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// Reserved prefix for database and collection names.
	reservedPrefix = "_ferretdb_"

	// Reserved prefix for SQLite table name.
	reservedTablePrefix = "sqlite"

	// metadataTableName is a SQLite table name where FerretDB metadata is stored.
	metadataTableName = reservedPrefix + "collections"
)

var (
	// ErrInvalidCollectionName indicates that a collection didn't pass name checks.
	ErrInvalidCollectionName = fmt.Errorf("invalid FerretDB collection name")
)

// Registry provides access to SQLite databases and collections information.
type Registry struct {
	p *pool.Pool
	l *zap.Logger
}

// NewRegistry creates a registry for the given directory.
func NewRegistry(dir string, l *zap.Logger) (*Registry, error) {
	p, err := pool.New(dir, l.Named("pool"))
	if err != nil {
		return nil, lazyerrors.Error(err)
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

// collectionToTable converts FerretDB collection name to SQLite table name.
// It returns error if the name is incorrect for SQLite backend.
func collectionToTable(collectionName string) (string, error) {
	if strings.HasPrefix(collectionName, reservedPrefix) {
		return "", ErrInvalidCollectionName
	}

	hash32 := fnv.New32a()
	must.NotFail(hash32.Write([]byte(collectionName)))

	// SQLite table cannot start with _sqlite prefix
	if strings.HasPrefix(strings.ToLower(collectionName), reservedTablePrefix) {
		collectionName = "_" + collectionName
	}
	// TODO case insensitive

	return collectionName + "_" + hex.EncodeToString(hash32.Sum(nil)), nil
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
// If collection name doesn't pass backend level validation it returns ErrInvalidCollectionName.
func (r *Registry) CollectionCreate(ctx context.Context, dbName string, collectionName string) (bool, error) {
	db, err := r.DatabaseGetOrCreate(ctx, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	tableName, err := collectionToTable(collectionName)
	if err != nil {
		return false, err
	}

	// TODO use transactions
	// https://github.com/FerretDB/FerretDB/issues/2747

	query := fmt.Sprintf("CREATE TABLE %q (_ferretdb_sjson TEXT)", tableName)
	if _, err = db.ExecContext(ctx, query); err != nil {
		var e *sqlite.Error
		if errors.As(err, &e) && e.Code() == sqlitelib.SQLITE_ERROR {
			return false, nil
		}

		return false, lazyerrors.Error(err)
	}

	query = fmt.Sprintf("INSERT INTO %q (name, table_name) VALUES (?, ?)", metadataTableName)
	if _, err = db.ExecContext(ctx, query, collectionName, tableName); err != nil {
		_, _ = db.ExecContext(ctx, fmt.Sprintf("DROP TABLE %q", tableName))
		return false, lazyerrors.Error(err)
	}

	return true, nil
}

func (r *Registry) CollectionGet(ctx context.Context, dbName string, collectionName string) (string, error) {
	db := r.p.GetExisting(ctx, dbName)
	if db == nil {
		return "", nil
	}

	query := fmt.Sprintf("SELECT table_name FROM %q WHERE name = ?", metadataTableName)
	rows, err := db.QueryContext(ctx, query, collectionName)
	if err != nil {
		return "", lazyerrors.Error(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", lazyerrors.New("no collection found")
	}

	var name string
	if err = rows.Scan(&name); err != nil {
		return "", lazyerrors.Error(err)
	}

	return name, nil
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

	tableName, err := r.CollectionGet(ctx, dbName, collectionName)
	if err != nil {
		return false, nil
	}

	// TODO use transactions
	// https://github.com/FerretDB/FerretDB/issues/2747

	query := fmt.Sprintf("DELETE FROM %q WHERE name = ?", metadataTableName)
	if _, err := db.ExecContext(ctx, query, collectionName); err != nil {
		return false, lazyerrors.Error(err)
	}

	query = fmt.Sprintf("DROP TABLE %q", tableName)
	if _, err := db.ExecContext(ctx, query); err != nil {
		return false, lazyerrors.Error(err)
	}

	return true, nil
}
