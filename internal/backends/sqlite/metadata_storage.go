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

	"golang.org/x/exp/slices"

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

// newMetadataStorage checks that dir is not empty,
// creates instance of metadataStorage and populates it with databases info.
func newMetadataStorage(dbPath string, pool *connPool) (*metadataStorage, error) {
	if dbPath == "" {
		return nil, errors.New("db path is empty")
	}

	storage := metadataStorage{
		connPool: pool,
		dir:      dbPath,
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
	dir      string
	connPool *connPool
}

// listDatabases list database names.
func (m *metadataStorage) listDatabases() ([]string, error) {
	var dbs []string

	err := filepath.WalkDir(m.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(d.Name(), dbExtension) {
			dbName, _ := strings.CutSuffix(d.Name(), dbExtension)

			dbs = append(dbs, dbName)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return dbs, nil
}

// listCollections list collection names for given database.
func (m *metadataStorage) listCollections(ctx context.Context, database string) ([]string, error) {
	exists, err := m.dbExists(database)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errDatabaseNotFound
	}

	colls, err := m.loadCollections(ctx, database)
	if err != nil {
		return nil, err
	}

	return colls, nil
}

// createDatabase adds database to metadata storage.
// It doesn't create database file.
func (m *metadataStorage) createDatabase(ctx context.Context, database string) error {
	exists, err := m.dbExists(database)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	conn, err := m.connPool.DB(database)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (sjson TEXT)", dbMetadataTableName)

	_, err = conn.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

// createCollection saves collection metadata to database file.
func (m *metadataStorage) createCollection(ctx context.Context, database, collection string) (string, error) {
	tableName, err := m.collectionInfo(ctx, database, collection)
	if err != nil {
		return "", err
	}

	conn, err := m.connPool.DB(database)
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

	return tableName, nil
}

// collectionInfo returns table name for given database name and collection name.
func (m *metadataStorage) collectionInfo(ctx context.Context, dbName, collName string) (string, error) {
	exists, err := m.dbExists(dbName)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", errDatabaseNotFound
	}

	conn, err := m.connPool.DB(dbName)
	if err != nil {
		return "", err
	}

	query := fmt.Sprintf("SELECT sjson FROM %s WHERE json_extract(sjson, '$.collection') = ?", dbMetadataTableName)

	rows, err := conn.QueryContext(ctx, query, collName)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var bytes []byte

	if !rows.Next() {
		return "", errCollectionNotFound
	}

	err = rows.Scan(&bytes)
	if err != nil {
		return "", err
	}

	doc, err := sjson.Unmarshal(bytes)
	if err != nil {
		return "", err
	}

	// TODO: proper error handling
	table := must.NotFail(doc.Get("table")).(string)

	return table, nil
}

// loadCollections loads collections metadata from database file.
func (m *metadataStorage) loadCollections(ctx context.Context, dbPath string) ([]string, error) {
	conn, err := m.connPool.DB(dbPath)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT sjson FROM %s", dbMetadataTableName)

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	// TODO: check error
	defer rows.Close()

	var collections []string

	for rows.Next() {
		var rawBytes []byte

		err = rows.Scan(rawBytes)
		if err != nil {
			return nil, err
		}

		doc, err := sjson.Unmarshal(rawBytes)
		if err != nil {
			return nil, errors.New("failed to unmarshal collection metadata")
		}

		// TODO: proper errors check
		collName := must.NotFail(doc.Get("collection")).(string)

		collections = append(collections, collName)
	}

	return collections, nil
}

// removeCollection removes collection metadata.
// It does not remove collection from database.
func (m *metadataStorage) removeCollection(ctx context.Context, database, collection string) error {
	exists, err := m.dbExists(database)
	if err != nil {
		return err
	}

	if !exists {
		return errDatabaseNotFound
	}

	colls, err := m.loadCollections(ctx, database)
	if err != nil {
		return err
	}

	// collection not found, nothing to do
	if !slices.Contains(colls, collection) {
		return nil
	}

	conn, err := m.connPool.DB(database)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE json_extract(sjson, '$.collection') = ?", dbMetadataTableName)

	res, err := conn.ExecContext(ctx, query, collection)
	if err != nil {
		return err
	}

	deleted, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if deleted != 1 {
		return errors.New("failed to remove collection")
	}

	return nil
}

func (m *metadataStorage) dbExists(database string) (bool, error) {
	dbs, err := m.listDatabases()
	if err != nil {
		return false, err
	}

	return slices.Contains(dbs, database), nil
}
