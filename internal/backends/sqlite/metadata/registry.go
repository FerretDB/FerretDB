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
	"encoding/hex"
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
type Registry struct {
	p *pool.Pool
	l *zap.Logger

	rw    sync.RWMutex
	colls map[string]map[string]*Collection // database name -> collection name -> collection
}

// NewRegistry creates a registry for SQLite databases in the directory specified by SQLite URI.
func NewRegistry(u string, l *zap.Logger) (*Registry, error) {
	p, err := pool.New(u, l)
	if err != nil {
		return nil, err
	}

	r := &Registry{
		p:     p,
		l:     l,
		colls: map[string]map[string]*Collection{},
	}

	for name, db := range p.DBS() {
		if err = r.loadCollections(context.Background(), name, db); err != nil {
			p.Close()
			return nil, lazyerrors.Error(err)
		}
	}

	return r, nil
}

// Close closes the registry.
func (r *Registry) Close() {
	r.p.Close()
}

// loadCollections gets collections metadata from the database during initialization.
func (r *Registry) loadCollections(ctx context.Context, dbName string, db *fsql.DB) error {
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

// getCollections returns collections metadata for the given database.
func (r *Registry) getCollections(ctx context.Context, dbName string, db *fsql.DB) map[string]*Collection {
	r.rw.RLock()
	colls := maps.Clone(r.colls[dbName])
	r.rw.RUnlock()

	if colls == nil {
		colls = map[string]*Collection{}
	}

	return colls
}

// DatabaseList returns a sorted list of existing databases.
func (r *Registry) DatabaseList(ctx context.Context) []string {
	return r.p.List(ctx)
}

// DatabaseGetExisting returns a connection to existing database or nil if it doesn't exist.
func (r *Registry) DatabaseGetExisting(ctx context.Context, dbName string) *fsql.DB {
	return r.p.GetExisting(ctx, dbName)
}

// DatabaseGetOrCreate returns a connection to existing database or newly created database.
func (r *Registry) DatabaseGetOrCreate(ctx context.Context, dbName string) (*fsql.DB, error) {
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
		r.DatabaseDrop(ctx, dbName)
		return nil, lazyerrors.Error(err)
	}

	return db, nil
}

// DatabaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
func (r *Registry) DatabaseDrop(ctx context.Context, dbName string) bool {
	r.rw.Lock()
	delete(r.colls, dbName)
	r.rw.Unlock()

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

	colls := r.getCollections(ctx, dbName, db)
	res := maps.Keys(colls)
	sort.Strings(res)
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

	colls := r.getCollections(ctx, dbName, db)
	if colls[collectionName] != nil {
		return false, nil
	}

	h := fnv.New32a()
	must.NotFail(h.Write([]byte(collectionName)))

	tableName := strings.ToLower(collectionName) + "_" + hex.EncodeToString(h.Sum(nil))
	if strings.HasPrefix(tableName, reservedTablePrefix) {
		tableName = "_" + tableName
	}

	// use transaction
	// TODO https://github.com/FerretDB/FerretDB/issues/2747

	q := fmt.Sprintf("CREATE TABLE %q (%s TEXT) STRICT", tableName, DefaultColumn)
	if _, err = db.ExecContext(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	q = fmt.Sprintf("CREATE UNIQUE INDEX %q ON %q (%s)", tableName+"_id", tableName, IDColumn)
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

	r.rw.Lock()

	if r.colls[dbName] == nil {
		r.colls[dbName] = map[string]*Collection{}
	}
	r.colls[dbName][collectionName] = c

	r.rw.Unlock()

	return true, nil
}

// CollectionGet returns collection metadata.
//
// If database or collection does not exist, nil is returned.
func (r *Registry) CollectionGet(ctx context.Context, dbName string, collectionName string) *Collection {
	db := r.p.GetExisting(ctx, dbName)
	if db == nil {
		return nil
	}

	colls := r.getCollections(ctx, dbName, db)

	return colls[collectionName]
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

	meta := r.CollectionGet(ctx, dbName, collectionName)
	if meta == nil {
		return false, nil
	}

	// use transaction
	// TODO https://github.com/FerretDB/FerretDB/issues/2747

	q := fmt.Sprintf("DELETE FROM %q WHERE name = ?", metadataTableName)
	if _, err := db.ExecContext(ctx, q, collectionName); err != nil {
		return false, lazyerrors.Error(err)
	}

	q = fmt.Sprintf("DROP TABLE %q", meta.TableName)
	if _, err := db.ExecContext(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	r.rw.Lock()

	delete(r.colls[dbName], collectionName)

	r.rw.Unlock()

	return true, nil
}

// Describe implements prometheus.Collector.
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(r, ch)
}

// Collect implements prometheus.Collector.
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	r.p.Collect(ch)

	r.rw.RLock()
	defer r.rw.RLock()

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
