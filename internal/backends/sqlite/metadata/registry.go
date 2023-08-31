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
	"sort"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata/pool"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
)

const (
	// This prefix is reserved by SQLite for internal use,
	// see https://www.sqlite.org/lang_createtable.html.
	reservedTablePrefix = "sqlite_"

	// SQLite table name where FerretDB metadata is stored.
	metadataTableName = "_ferretdb_collections"
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "sqlite_metadata"
)

// Registry provides access to SQLite databases and collections information.
//
// Exported methods are safe for concurrent use. Unexported methods are not.
type Registry struct {
	p *pool.Pool
	l *zap.Logger

	// rw protects colls but also acts like a global lock for the whole registry.
	// The latter effectively replaces transactions (see the sqlite backend description for more info).
	// One global lock should be replaced by more granular locks – one per database or even one per collection.
	// But that requires some redesign.
	// TODO https://github.com/FerretDB/FerretDB/issues/2755
	rw    sync.RWMutex
	colls map[string]map[string]*Collection // database name -> collection name -> collection
}

// NewRegistry creates a registry for SQLite databases in the directory specified by SQLite URI.
func NewRegistry(u string, l *zap.Logger) (*Registry, error) {
	p, initDBs, err := pool.New(u, l)
	if err != nil {
		return nil, err
	}

	r := &Registry{
		p:     p,
		l:     l,
		colls: map[string]map[string]*Collection{},
	}

	for name, db := range initDBs {
		if err = r.initCollections(context.Background(), name, db); err != nil {
			r.Close()
			return nil, lazyerrors.Error(err)
		}
	}

	return r, nil
}

// Close closes the registry.
func (r *Registry) Close() {
	r.p.Close()
}

// initCollections loads collections metadata from the database during initialization.
func (r *Registry) initCollections(ctx context.Context, dbName string, db *fsql.DB) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT name, table_name, settings FROM %q", metadataTableName))
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer rows.Close()

	colls := map[string]*Collection{}

	for rows.Next() {
		var c Collection
		if err = rows.Scan(&c.Name, &c.TableName, &c.Settings); err != nil {
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
func (r *Registry) DatabaseList(ctx context.Context) []string {
	defer observability.FuncCall(ctx)()

	return r.p.List(ctx)
}

// DatabaseGetExisting returns a connection to existing database or nil if it doesn't exist.
func (r *Registry) DatabaseGetExisting(ctx context.Context, dbName string) *fsql.DB {
	defer observability.FuncCall(ctx)()

	return r.p.GetExisting(ctx, dbName)
}

// databaseGetOrCreate returns a connection to existing database or newly created database.
//
// It does not hold the lock.
func (r *Registry) databaseGetOrCreate(ctx context.Context, dbName string) (*fsql.DB, error) {
	defer observability.FuncCall(ctx)()

	db, created, err := r.p.GetOrCreate(ctx, dbName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !created {
		return db, nil
	}

	q := fmt.Sprintf(
		"CREATE TABLE %q ("+
			"name TEXT NOT NULL UNIQUE CHECK(name != ''), "+
			"table_name TEXT NOT NULL UNIQUE CHECK(table_name != ''), "+
			"settings TEXT NOT NULL CHECK(settings != '')"+
			") STRICT",
		metadataTableName,
	)
	if _, err = db.ExecContext(ctx, q); err != nil {
		r.databaseDrop(ctx, dbName)
		return nil, lazyerrors.Error(err)
	}

	return db, nil
}

// DatabaseGetOrCreate returns a connection to existing database or newly created database.
func (r *Registry) DatabaseGetOrCreate(ctx context.Context, dbName string) (*fsql.DB, error) {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.databaseGetOrCreate(ctx, dbName)
}

// databaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
//
// It does not hold the lock.
func (r *Registry) databaseDrop(ctx context.Context, dbName string) bool {
	defer observability.FuncCall(ctx)()

	delete(r.colls, dbName)

	return r.p.Drop(ctx, dbName)
}

// DatabaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
func (r *Registry) DatabaseDrop(ctx context.Context, dbName string) bool {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	delete(r.colls, dbName)

	return r.p.Drop(ctx, dbName)
}

// CollectionList returns a sorted list of collections in the database.
//
// If database does not exist, no error is returned.
func (r *Registry) CollectionList(ctx context.Context, dbName string) ([]*Collection, error) {
	defer observability.FuncCall(ctx)()

	db := r.p.GetExisting(ctx, dbName)
	if db == nil {
		return nil, nil
	}

	r.rw.RLock()

	res := maps.Values(r.colls[dbName])

	r.rw.RUnlock()

	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })

	return res, nil
}

// CollectionCreate creates a collection in the database.
//
// Returned boolean value indicates whether the collection was created.
// If collection already exists, (false, nil) is returned.
func (r *Registry) CollectionCreate(ctx context.Context, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	r.rw.Lock()
	defer r.rw.Unlock()

	db, err := r.databaseGetOrCreate(ctx, dbName)
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

	// TODO https://github.com/FerretDB/FerretDB/issues/2760

	tableName := fmt.Sprintf("%s_%08x", strings.ToLower(collectionName), s)
	if strings.HasPrefix(tableName, reservedTablePrefix) {
		tableName = "_" + tableName
	}

	q := fmt.Sprintf("CREATE TABLE %[1]q (%[2]s TEXT NOT NULL CHECK(%[2]s != '')) STRICT", tableName, DefaultColumn)
	if _, err = db.ExecContext(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	pkName := tableName + "_id"
	q = fmt.Sprintf("CREATE UNIQUE INDEX %q ON %q (%s)", pkName, tableName, IDColumn)
	if _, err = db.ExecContext(ctx, q); err != nil {
		_, _ = db.ExecContext(ctx, fmt.Sprintf("DROP TABLE %q", tableName))
		return false, lazyerrors.Error(err)
	}

	c := &Collection{
		Name:      collectionName,
		TableName: tableName,
		Settings:  "{}",
	}

	q = fmt.Sprintf("INSERT INTO %q (name, table_name, settings) VALUES (?, ?, ?)", metadataTableName)
	if _, err = db.ExecContext(ctx, q, c.Name, c.TableName, c.Settings); err != nil {
		_, _ = db.ExecContext(ctx, fmt.Sprintf("DROP TABLE %q", tableName))
		return false, lazyerrors.Error(err)
	}

	if r.colls[dbName] == nil {
		r.colls[dbName] = map[string]*Collection{}
	}
	r.colls[dbName][collectionName] = c

	return true, nil
}

// CollectionGet returns collection metadata.
//
// If database or collection does not exist, nil is returned.
func (r *Registry) CollectionGet(ctx context.Context, dbName, collectionName string) *Collection {
	defer observability.FuncCall(ctx)()

	r.rw.RLock()
	defer r.rw.RUnlock()

	colls := r.colls[dbName]
	if colls == nil {
		return nil
	}

	return colls[collectionName]
}

// CollectionDrop drops a collection in the database.
//
// Returned boolean value indicates whether the collection was dropped.
// If database or collection did not exist, (false, nil) is returned.
func (r *Registry) CollectionDrop(ctx context.Context, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	db := r.p.GetExisting(ctx, dbName)
	if db == nil {
		return false, nil
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	colls := r.colls[dbName]
	if colls == nil {
		return false, nil
	}

	c := colls[collectionName]
	if c == nil {
		return false, nil
	}

	q := fmt.Sprintf("DELETE FROM %q WHERE name = ?", metadataTableName)
	if _, err := db.ExecContext(ctx, q, collectionName); err != nil {
		return false, lazyerrors.Error(err)
	}

	q = fmt.Sprintf("DROP TABLE %q", c.TableName)
	if _, err := db.ExecContext(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	delete(r.colls[dbName], collectionName)

	return true, nil
}

// CollectionRename renames a collection in the database.
//
// Returned boolean value indicates whether the collection was renamed.
// If database or collection did not exist, (false, nil) is returned.
func (r *Registry) CollectionRename(ctx context.Context, dbName, oldCollectionName, newCollectionName string) (bool, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/2760
	panic("not implemented")
}

// Describe implements prometheus.Collector.
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(r, ch)
}

// Collect implements prometheus.Collector.
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	r.p.Collect(ch)

	r.rw.RLock()
	defer r.rw.RUnlock()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "databases"),
			"The current number of database in the registry.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(len(r.colls)),
	)

	for db, colls := range r.colls {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, "collections"),
				"The current number of collections in the registry.",
				[]string{"db"}, nil,
			),
			prometheus.GaugeValue,
			float64(len(colls)),
			db,
		)
	}
}

// check interfaces
var (
	_ prometheus.Collector = (*Registry)(nil)
)
