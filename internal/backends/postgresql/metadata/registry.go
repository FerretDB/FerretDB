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
	"hash/fnv"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata/pool"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

const (
	// Reserved prefix for database and collection names.
	reservedPrefix = "_ferretdb_"

	// PostgreSQL table name where FerretDB metadata is stored.
	metadataTableName = reservedPrefix + "database_metadata"
)

// Registry provides access to PostgreSQL databases and collections information.
//
// Exported methods are safe for concurrent use. Unexported methods are not.
//
// All methods should call [getPool] to check authentication.
// There is no authorization yet – if username/password combination is correct,
// all databases and collections are visible as far as Registry is concerned.
//
//nolint:vet // for readability
type Registry struct {
	p *pool.Pool
	l *zap.Logger

	// rw protects colls but also acts like a global lock for the whole registry.
	// The latter effectively replaces transactions (see the postgresql backend package description for more info).
	// One global lock should be replaced by more granular locks – one per database or even one per collection.
	// But that requires some redesign.
	// TODO https://github.com/FerretDB/FerretDB/issues/2755
	rw    sync.RWMutex
	colls map[string]map[string]*Collection // database name -> collection name -> collection
}

// NewRegistry creates a registry for PostgreSQL databases with a given base URI.
func NewRegistry(u string, l *zap.Logger, sp *state.Provider) (*Registry, error) {
	p, err := pool.New(u, l, sp)
	if err != nil {
		return nil, err
	}

	r := &Registry{
		p:     p,
		l:     l,
		colls: map[string]map[string]*Collection{},
	}

	return r, nil
}

// Close closes the registry.
func (r *Registry) Close() {
	r.p.Close()
}

// getPool returns a pool of connections to PostgreSQL database
// for the username/password combination in the context using [conninfo].
//
// All methods should use that method to check authentication.
func (r *Registry) getPool(ctx context.Context) (*pgxpool.Pool, error) {
	username, password := conninfo.Get(ctx).Auth()

	p, err := r.p.Get(username, password)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return p, nil
}

// DatabaseList returns a sorted list of existing databases.
func (r *Registry) DatabaseList(ctx context.Context) ([]string, error) {
	defer observability.FuncCall(ctx)()

	if _, err := r.getPool(ctx); err != nil {
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
func (r *Registry) DatabaseGetExisting(ctx context.Context, dbName string) (*pgxpool.Pool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.RLock()
	defer r.rw.RUnlock()

	return r.databaseGetExisting(ctx, p, dbName)
}

// databaseGetExisting returns a connection to existing database or nil if it doesn't exist.
//
// It does not hold the lock.
func (r *Registry) databaseGetExisting(ctx context.Context, p *pgxpool.Pool, dbName string) (*pgxpool.Pool, error) {
	db := r.colls[dbName]
	if db == nil {
		return nil, nil
	}

	return p, nil
}

// DatabaseGetOrCreate returns a connection to existing database or newly created database.
//
// If the user is not authenticated, it returns error.
func (r *Registry) DatabaseGetOrCreate(ctx context.Context, dbName string) (*pgxpool.Pool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.databaseGetOrCreate(ctx, p, dbName)
}

// databaseGetOrCreate returns a connection to existing database or newly created database.
//
// The dbName must be a validated database name.
//
// It does not hold the lock.
func (r *Registry) databaseGetOrCreate(ctx context.Context, p *pgxpool.Pool, dbName string) (*pgxpool.Pool, error) {
	defer observability.FuncCall(ctx)()

	db, err := r.databaseGetExisting(ctx, p, dbName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if db != nil {
		return db, nil
	}

	q := fmt.Sprintf(
		`CREATE SCHEMA IF NOT EXISTS %s`,
		pgx.Identifier{dbName}.Sanitize(),
	)

	if _, err = p.Exec(ctx, q); err != nil {
		return nil, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (%s jsonb)`,
		pgx.Identifier{dbName, metadataTableName}.Sanitize(),
		DefaultColumn,
	)

	if _, err = p.Exec(ctx, q); err != nil {
		_, _ = r.databaseDrop(ctx, p, dbName)
		return nil, lazyerrors.Error(err)
	}

	r.colls[dbName] = map[string]*Collection{}

	return p, nil
}

// DatabaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
//
// If the user is not authenticated, it returns error.
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

// DatabaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
//
// It does not hold the lock.
func (r *Registry) databaseDrop(ctx context.Context, p *pgxpool.Pool, dbName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	db, err := r.databaseGetExisting(ctx, p, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	if db == nil {
		return false, nil
	}

	q := fmt.Sprintf(
		`DROP SCHEMA %s CASCADE`,
		pgx.Identifier{dbName}.Sanitize(),
	)

	if _, err := p.Exec(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	delete(r.colls, dbName)

	return true, nil
}

// CollectionList returns a sorted copy of collections in the database.
//
// If database does not exist, no error is returned.
func (r *Registry) CollectionList(ctx context.Context, dbName string) ([]*Collection, error) {
	defer observability.FuncCall(ctx)()

	db, err := r.DatabaseGetExisting(ctx, dbName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if db == nil {
		return nil, nil
	}

	r.rw.RLock()

	res := make([]*Collection, 0, len(r.colls[dbName]))
	for _, c := range r.colls[dbName] {
		res = append(res, c.deepCopy())
	}

	r.rw.RUnlock()

	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })

	return res, nil
}

// CollectionCreate creates a collection in the database.
// Database will be created automatically if needed.
//
// Returned boolean value indicates whether the collection was created.
// If collection already exists, (false, nil) is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) CollectionCreate(ctx context.Context, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.collectionCreate(ctx, p, dbName, collectionName)
}

// collectionCreate creates a collection in the database.
// Database will be created automatically if needed.
//
// Returned boolean value indicates whether the collection was created.
// If collection already exists, (false, nil) is returned.
//
// It does not hold the lock.
func (r *Registry) collectionCreate(ctx context.Context, p *pgxpool.Pool, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.databaseGetOrCreate(ctx, p, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	colls := r.colls[dbName]
	if colls != nil && colls[collectionName] != nil {
		return false, nil
	}

	h := fnv.New32a()
	must.NotFail(h.Write([]byte(collectionName)))
	s := h.Sum32()

	var tableName string
	list := maps.Values(colls)

	for {
		tableName = fmt.Sprintf("%s_%08x", strings.ToLower(collectionName), s)
		if strings.HasPrefix(tableName, reservedPrefix) {
			tableName = "_" + tableName
		}

		if !slices.ContainsFunc(list, func(c *Collection) bool { return c.TableName == tableName }) {
			break
		}

		// table already exists, generate a new table name by incrementing the hash
		s++
	}

	q := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (%s jsonb)`,
		pgx.Identifier{dbName, tableName}.Sanitize(),
		DefaultColumn,
	)

	if _, err = p.Exec(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	c := &Collection{
		Name:      collectionName,
		TableName: tableName,
	}

	b, err := sjson.Marshal(c.Marshal())
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	// create PG index for collection name
	// TODO https://github.com/FerretDB/FerretDB/issues/3375

	// create PG index for table name
	// TODO https://github.com/FerretDB/FerretDB/issues/3375

	q = fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES ($1)`,
		pgx.Identifier{dbName, metadataTableName}.Sanitize(),
		DefaultColumn,
	)

	if _, err = p.Exec(ctx, q, string(b)); err != nil {
		q = fmt.Sprintf(`DROP TABLE %s`, pgx.Identifier{dbName, tableName}.Sanitize())
		_, _ = p.Exec(ctx, q)

		return false, lazyerrors.Error(err)
	}

	if r.colls[dbName] == nil {
		r.colls[dbName] = map[string]*Collection{}
	}
	r.colls[dbName][collectionName] = c

	return true, nil
}

// CollectionGet returns a copy of collection metadata.
// It can be safely modified by a caller.
//
// If database or collection does not exist, nil is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) CollectionGet(ctx context.Context, dbName, collectionName string) (*Collection, error) {
	defer observability.FuncCall(ctx)()

	if _, err := r.getPool(ctx); err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.RLock()
	defer r.rw.RUnlock()

	return r.collectionGet(dbName, collectionName), nil
}

// collectionGet returns a copy of collection metadata.
// It can be safely modified by a caller.
//
// If database or collection does not exist, nil is returned.
//
// It does not hold the lock.
func (r *Registry) collectionGet(dbName, collectionName string) *Collection {
	colls := r.colls[dbName]
	if colls == nil {
		return nil
	}

	return colls[collectionName].deepCopy()
}

// CollectionDrop drops a collection in the database.
//
// Returned boolean value indicates whether the collection was dropped.
// If database or collection did not exist, (false, nil) is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) CollectionDrop(ctx context.Context, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.collectionDrop(ctx, p, dbName, collectionName)
}

// collectionDrop drops a collection in the database.
//
// Returned boolean value indicates whether the collection was dropped.
// If database or collection did not exist, (false, nil) is returned.
//
// It does not hold the lock.
func (r *Registry) collectionDrop(ctx context.Context, p *pgxpool.Pool, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	db, err := r.databaseGetExisting(ctx, p, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	if db == nil {
		return false, nil
	}

	c := r.collectionGet(dbName, collectionName)
	if c == nil {
		return false, nil
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/811
	q := fmt.Sprintf(
		`DROP TABLE IF EXISTS %s CASCADE`,
		pgx.Identifier{dbName, c.TableName}.Sanitize(),
	)

	if _, err = p.Exec(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	arg, err := sjson.MarshalSingleValue(c.Name)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`DELETE FROM %s WHERE %s IN ($1)`,
		pgx.Identifier{dbName, metadataTableName}.Sanitize(),
		IDColumn,
	)

	if _, err := p.Exec(ctx, q, arg); err != nil {
		// the table has been dropped but metadata related to the table is out of sync
		return false, lazyerrors.Error(err)
	}

	delete(r.colls[dbName], collectionName)

	return true, nil
}

// CollectionRename renames a collection in the database.
//
// The collection name is update, but original table name is kept.
//
// Returned boolean value indicates whether the collection was renamed.
// If database or collection did not exist, (false, nil) is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) CollectionRename(ctx context.Context, dbName, oldCollectionName, newCollectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	db, err := r.databaseGetExisting(ctx, p, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	if db == nil {
		return false, nil
	}

	c := r.collectionGet(dbName, oldCollectionName)
	if c == nil {
		return false, nil
	}

	c.Name = newCollectionName

	b, err := sjson.Marshal(c.Marshal())
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	arg, err := sjson.MarshalSingleValue(oldCollectionName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	q := fmt.Sprintf(
		`UPDATE %s SET %s = ? WHERE %s = ?`,
		pgx.Identifier{dbName, metadataTableName}.Sanitize(),
		DefaultColumn,
		IDColumn,
	)

	if _, err := p.Exec(ctx, q, string(b), arg); err != nil {
		return false, lazyerrors.Error(err)
	}

	r.colls[dbName][newCollectionName] = c
	delete(r.colls[dbName], oldCollectionName)

	return true, nil
}
