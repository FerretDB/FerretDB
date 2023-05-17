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

package sqlite

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// Reserved prefix for database and collection names.
	reservedPrefix = "_ferretdb_"

	// Database metadata table name.
	dbMetadataTableName = reservedPrefix + "database_metadata"

	dbExtension = ".sqlite"
)

var (
	errDatabaseNotFound   = errors.New("database not found")
	errCollectionNotFound = errors.New("collection not found")
)

// newMetadataStorage returns instance of metadata storage.
func newMetadataStorage(dbPath string, pool *connPool) (*metadataStorage, error) {
	if dbPath == "" {
		return nil, errors.New("db path is empty")
	}

	return &metadataStorage{
		connPool: pool,
		dbPath:   dbPath,
		dbs:      map[string]*dbInfo{},
	}, nil
}

// metadataStorage provide access to database metadata.
// It uses connection pool to load and store metadata.
type metadataStorage struct { //nolint:vet // for readability
	dbPath string

	connPool *connPool

	mx  sync.Mutex
	dbs map[string]*dbInfo
}

type dbInfo struct {
	collections map[string]string
	// TODO: add indexes, etc.
}

// ListDatabases list database names.
func (m *metadataStorage) ListDatabases() ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.dbs != nil {
		return maps.Keys(m.dbs), nil
	}

	var dbs []string

	err := filepath.WalkDir(m.dbPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == dbExtension {
			dbs = append(dbs, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// fill in all databases that were found.
	for _, db := range dbs {
		m.dbs[db] = nil
	}

	return dbs, nil
}

// ListCollections list collection names for given database.
func (m *metadataStorage) ListCollections(ctx context.Context, database string) ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return nil, errors.New("database not found")
	}

	// no metadata about collections loaded, load now
	if db == nil {
		var err error

		db, err = m.load(ctx, database)
		if err != nil {
			return nil, err
		}

		m.dbs[database] = db
	}

	return maps.Keys(db.collections), nil
}

// RemoveDatabase removes database metadata.
// It does not remove database file.
func (m *metadataStorage) RemoveDatabase(database string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	_, ok := m.dbs[database]
	if !ok {
		return errors.New("database not found")
	}

	delete(m.dbs, database)

	return nil
}

// RemoveCollection removes collection metadata.
// It does not remove collection from database.
func (m *metadataStorage) RemoveCollection(database, collection string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return errors.New("database not found")
	}

	_, ok = db.collections[collection]
	if !ok {
		return errors.New("collection not found")
	}

	delete(db.collections, collection)

	return nil
}

// CreateDatabase adds database to metadata storage.
// It doesn't create database file.
func (m *metadataStorage) CreateDatabase(database string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	_, ok := m.dbs[database]
	if ok {
		return errors.New("database already exists")
	}

	m.dbs[database] = nil

	return nil
}

// CreateCollection saves collection metadata to database file.
func (m *metadataStorage) CreateCollection(ctx context.Context, database, collection string) (string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return "", errors.New("database not found")
	}

	_, ok = db.collections[collection]
	if ok {
		return "", errors.New("collection already exists")
	}

	tableName := tableNameFromCollectionName(collection)

	err := m.saveCollection(ctx, database, collection, tableName)
	if err != nil {
		return "", err
	}

	db.collections[collection] = tableName

	return tableName, nil
}

// load loads database metadata from database file.
func (m *metadataStorage) load(ctx context.Context, dbName string) (*dbInfo, error) {
	conn, err := m.connPool.DB(dbName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT sjson FROM %s", dbMetadataTableName)

	result, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	// TODO: check error
	defer result.Close()

	metadata := dbInfo{
		collections: map[string]string{},
	}

	for result.Next() {
		var rawBytes []byte

		err = result.Scan(rawBytes)
		if err != nil {
			return nil, err
		}

		doc, err := sjson.Unmarshal(rawBytes)
		if err != nil {
			return nil, errors.New("failed to unmarshal collection metadata")
		}

		// TODO: proper errors check
		collName := must.NotFail(doc.Get("collection")).(string)
		tableName := must.NotFail(doc.Get("table")).(string)

		metadata.collections[collName] = tableName
	}

	return &metadata, nil
}

// saveCollection saves collection metadata to database file.
func (m *metadataStorage) saveCollection(ctx context.Context, dbName, collName, tableName string) error {
	conn, err := m.connPool.DB(dbName)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO %s (sjson) VALUES (?)", dbMetadataTableName)

	bytes, err := sjson.Marshal(must.NotFail(types.NewDocument("collection", collName, "table", tableName)))
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, query, bytes)
	if err != nil {
		return err
	}

	return nil
}

// CollectionInfo returns table name for given database name and collection name.
func (m *metadataStorage) CollectionInfo(dbName string, collName string) (string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[dbName]
	if !ok {
		return "", errDatabaseNotFound
	}

	table, ok := db.collections[collName]
	if !ok {
		return "", errCollectionNotFound
	}

	return table, nil
}

// tableNameFromCollectionName mangles collection name to table name.
// TODO: implement proper mangle if needed.
func tableNameFromCollectionName(name string) string {
	return name
}
