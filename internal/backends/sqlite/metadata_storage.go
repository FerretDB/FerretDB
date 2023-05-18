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
	"strings"
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

// newMetadataStorage checks that dbPath is not empty,
// creates instance of metadataStorage and populates it with databases info.
func newMetadataStorage(dbPath string, pool *connPool) (*metadataStorage, error) {
	if dbPath == "" {
		return nil, errors.New("db path is empty")
	}

	storage := metadataStorage{
		connPool: pool,
		dbPath:   dbPath,
		dbs:      map[string]*dbInfo{},
		mx:       sync.Mutex{},
	}

	_, err := storage.listDatabases()
	if err != nil {
		return nil, err
	}

	return &storage, nil
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
	path        string
	collections map[string]string
	// TODO: add indexes, etc.
}

// listDatabases list database names.
func (m *metadataStorage) listDatabases() ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	dbs := map[string]*dbInfo{}

	err := filepath.WalkDir(m.dbPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(d.Name(), dbExtension) {
			dbName, _ := strings.CutSuffix(d.Name(), dbExtension)

			dbs[dbName] = &dbInfo{
				path:        path,
				collections: nil,
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	m.dbs = dbs

	return maps.Keys(dbs), nil
}

// listCollections list collection names for given database.
func (m *metadataStorage) listCollections(ctx context.Context, database string) ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return nil, errDatabaseNotFound
	}

	// no metadata about collections loaded, load it
	if db.collections == nil {
		var err error

		db, err = m.loadCollections(ctx, db.path)
		if err != nil {
			return nil, err
		}

		m.dbs[database] = db
	}

	return maps.Keys(db.collections), nil
}

// createDatabase adds database to metadata storage.
// It doesn't create database file.
func (m *metadataStorage) createDatabase(ctx context.Context, database string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	// database already exists
	if ok {
		return nil
	}

	db = &dbInfo{
		path: filepath.Join(m.dbPath, database+dbExtension),
	}

	conn, err := m.connPool.DB(db.path)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (sjson TEXT)", dbMetadataTableName)

	_, err = conn.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	m.dbs[database] = db

	return nil
}

// createCollection saves collection metadata to database file.
func (m *metadataStorage) createCollection(ctx context.Context, database, collection string) (string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return "", errDatabaseNotFound
	}

	tableName, ok := db.collections[collection]
	if ok {
		return tableName, nil
	}

	if db.collections == nil {
		db.collections = map[string]string{}
	}

	// TODO: transform table name if needed
	tableName = collection

	conn, err := m.connPool.DB(db.path)
	if err != nil {
		return "", err
	}

	tableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (sjson TEXT)", tableName)

	_, err = conn.ExecContext(ctx, tableQuery)
	if err != nil {
		return "", err
	}

	query := fmt.Sprintf("INSERT INTO %s (sjson) VALUES (?)", dbMetadataTableName)

	bytes, err := sjson.Marshal(must.NotFail(types.NewDocument("collection", collection, "table", tableName)))
	if err != nil {
		return "", err
	}

	_, err = conn.ExecContext(ctx, query, bytes)
	if err != nil {
		return "", err
	}

	db.collections[collection] = tableName

	return tableName, nil
}

func (m *metadataStorage) databasePath(database string) (string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return "", errDatabaseNotFound
	}

	return db.path, nil
}

// collectionInfo returns table name for given database name and collection name.
func (m *metadataStorage) collectionInfo(dbName string, collName string) (string, error) {
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

// removeDatabase removes database metadata.
// It does not remove database file.
func (m *metadataStorage) removeDatabase(database string) bool {
	m.mx.Lock()
	defer m.mx.Unlock()

	_, ok := m.dbs[database]
	if !ok {
		return false
	}

	delete(m.dbs, database)

	return true
}

// loadCollections loads collections metadata from database file.
func (m *metadataStorage) loadCollections(ctx context.Context, dbPath string) (*dbInfo, error) {
	conn, err := m.connPool.DB(dbPath)
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

// removeCollection removes collection metadata.
// It does not remove collection from database.
func (m *metadataStorage) removeCollection(database, collection string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return errDatabaseNotFound
	}

	_, ok = db.collections[collection]
	if !ok {
		return errCollectionNotFound
	}

	delete(db.collections, collection)

	return nil
}
