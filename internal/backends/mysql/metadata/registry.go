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

package metadata

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/mysql/metadata/pool"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

const (
	// MySQL table name where FerretDB metadata is stored.
	metadataTableName = backends.ReservedPrefix + "database_metadata"

	// MySQL max table name length.
	maxTableNameLength = 64
)

// Registry provides access to MySQL databases and collections information
//
// Exported methods are safe for concurrent use. Unexported methods are not.
type Registry struct {
	p *pool.Pool
	l *zap.Logger

	// rw protects colls but also acts like a global lock for the whole registry.
	// The latter effectively replaces transactions (see the mysql backend package description for more info).
	// One global lock should be replaced by more granular locks – one per database or even one per collection.
	// But that requires some redesign.
	// TODO https://github.com/FerretDB/FerretDB/issues/2755
	rw    sync.RWMutex
	colls map[string]map[string]*Collection // database name -> collection name -> collection
}

// NewRegistry creates a registry for mysql databases with a given base URI.
func NewRegistry(u string, l *zap.Logger, sp *state.Provider) (*Registry, error) {
	p, err := pool.New(u, l, sp)
	if err != nil {
		return nil, err
	}

	r := &Registry{
		p: p,
		l: l,
	}

	return r, nil
}

// Close closes the registry.
func (r *Registry) Close() {
	r.p.Close()
}

// getPool returns a pool of connection to MySQL database
// for the username/password combination in the context using [conninfo].
//
// It loads metadata if it hasn't been loaded from the database yet.
//
// It acquires read lock to check metadata, if metadata is empty it acquires write lock
// to load metadata, so it is safe for concurrent use.
//
// All methods should use this method to check authentication and load metadata.
func (r *Registry) getPool(ctx context.Context) (*fsql.DB, error) {
	username, password := conninfo.Get(ctx).Auth()

	p, err := r.p.Get(username, password)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.RLock()
	if r.colls != nil {
		r.rw.RUnlock()
		return p, nil
	}
	r.rw.RUnlock()

	r.rw.Lock()
	defer r.rw.Unlock()

	dbNames, err := r.initDBs(ctx, p)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.colls = make(map[string]map[string]*Collection, len(dbNames))
	return p, nil
}

// initDBs returns a list of database names using schema information.
// It fetches existing schema (excluding ones reserved for MySQL),
// then finds and returns the schema that contains FerretDB metadata table.
func (r *Registry) initDBs(ctx context.Context, db *fsql.DB) ([]string, error) {
	q := strings.TrimSpace(`
		SELECT schema_name 
		FROM information_schema.schemata
	`)

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var dbNames []string

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, lazyerrors.Error(err)
		}

		// schema created by MySQL can be used as a FerretDB database,
		// but if it does not contain FerretDB metadata table, it is not used by FerretDB
		q := strings.TrimSpace(`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = ? AND table_name = ?
			)
		`)

		var exist bool
		if err := db.QueryRowContext(ctx, q, dbName, metadataTableName).Scan(&exist); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if exist {
			dbNames = append(dbNames, dbName)
		}
	}

	return dbNames, nil
}

// initCollection loads collection metadata from the database during initialization.
func (r *Registry) initCollections(ctx context.Context, dbName string, db *fsql.DB) error {
	defer observability.FuncCall(ctx)()

	q := fmt.Sprintf(
		`SELECT %s FROM %s.%s`,
		DefaultColumn,
		strings.TrimSpace(dbName),
		metadataTableName,
	)

	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer rows.Close()

	colls := map[string]*Collection{}

	for rows.Next() {
		var c Collection

		if err := rows.Scan(&c); err != nil {
			return lazyerrors.Error(err)
		}

		colls[c.Name] = &c
	}

	if err = rows.Err(); err != nil {
		return lazyerrors.Error(err)
	}

	r.colls[dbName] = colls

	return nil
}

// DatabaseList returns a sorted list of existing databases.
//
// If the user is not authenticated, it returns error.
func (r *Registry) DatabaseList(ctx context.Context) ([]string, error) {
	defer observability.FuncCall(ctx)()

	_, err := r.getPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.RLock()
	defer r.rw.RUnlock()

	res := maps.Keys(r.colls)
	sort.Strings(res)

	return res, nil
}

// DatabaseGetExisting returns a connection to existing database or nil if it doesn't exist.
//
// If the user is not authenticated, it returns error.
func (r *Registry) DatabaseGetExisting(ctx context.Context, dbName string) (*fsql.DB, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.RLock()
	defer r.rw.RUnlock()

	db := r.colls[dbName]
	if db == nil {
		return nil, nil
	}

	return p, nil
}

// DatabaseGetOrCreate returns a connection to existing database or newly created database.
//
// The dbName must be a validated database name.
//
// If the user is not authenticated, it returns an error.
func (r *Registry) DatabaseGetOrCreate(ctx context.Context, dbName string) (*fsql.DB, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.databaseGetOrCreate(ctx, p, dbName)
}

// databaseGetOrCreate returns a connection to existing database or a newly created database.
//
// The dbName must be a validated database name.
//
// It does not hold the lock
func (r *Registry) databaseGetOrCreate(ctx context.Context, p *fsql.DB, dbName string) (*fsql.DB, error) {
	defer observability.FuncCall(ctx)()

	db := r.colls[dbName]
	if db != nil {
		return p, nil
	}

	q := fmt.Sprintf(
		`CREATE SCHEMA %s`,
		dbName,
	)

	var err error
	if _, err = p.ExecContext(ctx, q); err != nil {
		return nil, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`CREATE TABLE %s.%s (%s json)`,
		dbName,
		metadataTableName,
		DefaultColumn,
	)

	if _, err := p.ExecContext(ctx, q); err != nil {
		_, _ = r.databaseDrop(ctx, p, dbName)
		return nil, lazyerrors.Error(err)
	}

	// json columns cannot be indexed directly in MySQL. A workaround
	// around this is done by creating a generated column that extracts
	// information that should be indexed.
	//
	// https://dev.mysql.com/doc/refman/5.7/en/create-table-secondary-indexes.html#json-column-indirect-index

	q = fmt.Sprintf(
		`ALTER TABLE %s.%s
		 ADD COLUMN %s VARCHAR(255) GENERATED ALWAYS AS ((%s)) STORED,
		 ADD UNIQUE INDEX %s (%s)
		`,
		dbName,
		metadataTableName,
		TableIdxColumn+"_id",
		IDColumn,
		metadataTableName+"_id_idx",
		TableIdxColumn+"_id",
	)

	if _, err := p.ExecContext(ctx, q); err != nil {
		_, _ = r.databaseDrop(ctx, p, dbName)
		return nil, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`ALTER TABLE %s.%s
		 ADD COLUMN %s VARCHAR(255) GENERATED ALWAYS AS ((%s->'table')) STORED,
		 ADD UNIQUE INDEX %s (%s)
		`,
		dbName,
		metadataTableName,
		TableIdxColumn,
		DefaultColumn,
		metadataTableName+"_table_idx",
		TableIdxColumn,
	)

	if _, err = p.ExecContext(ctx, q); err != nil {
		_, _ = r.databaseDrop(ctx, p, dbName)
		return nil, lazyerrors.Error(err)
	}

	r.colls[dbName] = map[string]*Collection{}

	return p, nil
}

// DatabaseDrop drops the database
//
// Returned boolean value indicates whether the database was dropped.
// If database does not exist, (false, nil) is returned.
//
// If user is not authenticated, it returns error.
func (r *Registry) DatabaseDrop(ctx context.Context, dbName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.databaseDrop(ctx, p, dbName)
}

// DatabaseDrop drops the database
//
// Returned boolean value indicates whether the database was dropped.
// If database does not exist, (false, nil) is returned.
//
// It does not hold the lock
func (r *Registry) databaseDrop(ctx context.Context, p *fsql.DB, dbName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	db := r.colls[dbName]
	if db == nil {
		return false, nil
	}

	// TODO: fix cascade delete for mysql
	q := fmt.Sprintf(
		`DROP DATABASE %s`,
		dbName,
	)

	if _, err := p.ExecContext(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	delete(r.colls, dbName)

	return true, nil
}
