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

	r.rw.RLock()
	defer r.rw.RUnlock()

	p, err := r.getPool(ctx)
	if err != nil {
		return nil, err
	}

	dbs := r.colls[dbName]
	if dbs == nil {
		p.Close()
		return nil, nil
	}

	return p, nil
}

// DatabaseGetOrCreate returns a connection to existing database or newly created database.
func (r *Registry) DatabaseGetOrCreate(ctx context.Context, dbName string) (*pgxpool.Pool, error) {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.databaseGetOrCreate(ctx, dbName)
}

// databaseGetOrCreate returns a connection to existing database or newly created database.
//
// The dbName must be a validated database name.
//
// It does not hold the lock.
func (r *Registry) databaseGetOrCreate(ctx context.Context, dbName string) (*pgxpool.Pool, error) {
	defer observability.FuncCall(ctx)()

	if p, err := r.DatabaseGetExisting(ctx, dbName); err != nil {
		return nil, lazyerrors.Error(err)
	} else if p != nil {
		return p, nil
	}

	p, err := r.getPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	q := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, pgx.Identifier{dbName}.Sanitize())

	if _, err = p.Exec(ctx, q); err != nil {
		return nil, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (%s jsonb)`,
		pgx.Identifier{dbName, metadataTableName}.Sanitize(),
		DefaultColumn,
	)

	if _, err = p.Exec(ctx, q); err != nil {
		r.databaseDrop(ctx, dbName)
		return nil, lazyerrors.Error(err)
	}

	r.colls[dbName] = map[string]*Collection{}

	return p, nil
}

// DatabaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
func (r *Registry) DatabaseDrop(ctx context.Context, dbName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.databaseDrop(ctx, dbName)
}

// DatabaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
//
// It does not hold the lock.
func (r *Registry) databaseDrop(ctx context.Context, dbName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.DatabaseGetExisting(ctx, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	if p == nil {
		return false, nil
	}

	p, err = r.getPool(ctx)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	q := fmt.Sprintf(`DROP SCHEMA %s CASCADE`, pgx.Identifier{dbName}.Sanitize())

	if _, err := p.Exec(ctx, q); err == nil {
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
func (r *Registry) CollectionCreate(ctx context.Context, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.collectionCreate(ctx, dbName, collectionName)
}

// collectionCreate creates a collection in the database.
// Database will be created automatically if needed.
//
// Returned boolean value indicates whether the collection was created.
// If collection already exists, (false, nil) is returned.
//
// It does not hold the lock.
func (r *Registry) collectionCreate(ctx context.Context, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.databaseGetOrCreate(ctx, dbName)
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

	q := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s jsonb)`, pgx.Identifier{dbName, tableName}.Sanitize(), DefaultColumn)
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

	q = fmt.Sprintf("INSERT INTO %s (%s) VALUES (?)", metadataTableName, DefaultColumn)
	if _, err = p.Exec(ctx, q, string(b)); err != nil {
		_, _ = p.Exec(ctx, fmt.Sprintf("DROP TABLE %s", pgx.Identifier{dbName, tableName}.Sanitize()))
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
func (r *Registry) CollectionGet(ctx context.Context, dbName, collectionName string) *Collection {
	defer observability.FuncCall(ctx)()

	r.rw.RLock()
	defer r.rw.RUnlock()

	return r.collectionGet(dbName, collectionName)
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
func (r *Registry) CollectionDrop(ctx context.Context, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.collectionDrop(ctx, dbName, collectionName)
}

// collectionDrop drops a collection in the database.
//
// Returned boolean value indicates whether the collection was dropped.
// If database or collection did not exist, (false, nil) is returned.
//
// It does not hold the lock.
func (r *Registry) collectionDrop(ctx context.Context, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.DatabaseGetExisting(ctx, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	if p == nil {
		return false, nil
	}

	c := r.collectionGet(dbName, collectionName)
	if c == nil {
		return false, nil
	}

	q := fmt.Sprintf(`DELETE FROM %s WHERE %s = ?`, metadataTableName, IDColumn)
	if _, err := p.Exec(ctx, q, c.Name); err != nil {
		return false, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/811
	q = fmt.Sprintf(`DROP TABLE IF EXISTS %s CASCADE` + pgx.Identifier{dbName, c.TableName}.Sanitize())
	if _, err = p.Exec(ctx, q); err != nil {
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
func (r *Registry) CollectionRename(ctx context.Context, dbName, oldCollectionName, newCollectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	p, err := r.DatabaseGetExisting(ctx, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	if p == nil {
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

	q := fmt.Sprintf(`UPDATE %s SET %s = ? WHERE %s = ?`, metadataTableName, DefaultColumn, IDColumn)
	if _, err := p.Exec(ctx, q, string(b), oldCollectionName); err != nil {
		return false, lazyerrors.Error(err)
	}

	r.colls[dbName][newCollectionName] = c
	delete(r.colls[dbName], oldCollectionName)

	return true, nil
}

// IndexesCreate creates indexes in the collection.
//
// Existing indexes with given names are ignored (TODO?).
func (r *Registry) IndexesCreate(ctx context.Context, dbName, collectionName string, indexes []IndexInfo) error {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.indexesCreate(ctx, dbName, collectionName, indexes)
}

// indexesCreate creates indexes in the collection.
//
// Existing indexes with given names are ignored (TODO?).
//
// It does not hold the lock.
func (r *Registry) indexesCreate(ctx context.Context, dbName, collectionName string, indexes []IndexInfo) error {
	defer observability.FuncCall(ctx)()

	_, err := r.DatabaseGetExisting(ctx, dbName)
	if err != nil {
		return lazyerrors.Error(err)
	}

	_, err = r.collectionCreate(ctx, dbName, collectionName)
	if err != nil {
		return lazyerrors.Error(err)
	}

	c := r.collectionGet(dbName, collectionName)
	if c == nil {
		panic("collection does not exist")
	}

	p, err := r.DatabaseGetExisting(ctx, dbName)
	if err != nil {
		return lazyerrors.Error(err)
	}

	created := make([]string, 0, len(indexes))

	for _, index := range indexes {
		if slices.ContainsFunc(c.Settings.Indexes, func(i IndexInfo) bool { return index.Name == i.Name }) {
			continue
		}

		q := "CREATE "

		if index.Unique {
			q += "UNIQUE "
		}

		q += "INDEX IF NOT EXISTS %q ON %q (%s)"

		var args []any

		columns := make([]string, len(index.Key))
		for i, key := range index.Key {
			order := "ASC"
			if key.Descending {
				order = "DESC"
			}

			var placeholders []string
			for _, k := range strings.Split(key.Field, ".") {
				args = append(args, k)
				placeholders = append(placeholders, "?")
			}

			columns[i] = fmt.Sprintf(`((%s->%s)) %s`, DefaultColumn, strings.Join(placeholders, " -> "), order)
		}

		q = fmt.Sprintf(
			q,
			pgx.Identifier{index.Name}.Sanitize(),
			pgx.Identifier{dbName, c.TableName}.Sanitize(),
			strings.Join(columns, ", "),
		)

		if _, err := p.Exec(ctx, q, args...); err != nil {
			// TODO drop index
			return lazyerrors.Error(err)
		}

		created = append(created, index.Name)
		c.Settings.Indexes = append(c.Settings.Indexes, index)
	}

	b, err := sjson.Marshal(c.Marshal())
	if err != nil {
		return lazyerrors.Error(err)
	}

	q := fmt.Sprintf(`UPDATE %s SET %s = ? WHERE %s = ?`, metadataTableName, DefaultColumn, IDColumn)
	if _, err := p.Exec(ctx, q, string(b), collectionName); err != nil {
		// todo drop index
		_ = created
		return lazyerrors.Error(err)
	}

	r.colls[dbName][collectionName] = c

	return nil
}

// IndexesDrop drops provided indexes for the given collection.
func (r *Registry) IndexesDrop(ctx context.Context, dbName, collectionName string, toDrop []string) error {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	panic("not implemented")
}
