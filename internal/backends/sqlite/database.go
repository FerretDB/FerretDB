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
	"fmt"
	"os"
	"path/filepath"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// database implements backends.Database interface.
type database struct {
	b    *backend
	name string
}

// newDatabase creates a new Database.
func newDatabase(b *backend, name string) backends.Database {
	return backends.DatabaseContract(&database{
		b:    b,
		name: name,
	})
}

// Close implements backends.Database interface.
func (db *database) Close() {
	db.b.pool.CloseDB(db.name) // TODO: Implement
}

// Collection implements backends.Database interface.
func (db *database) Collection(name string) backends.Collection {
	return newCollection(db, name)
}

// ListCollections implements backends.Database interface.
//
//nolint:lll // for readability
func (db *database) ListCollections(ctx context.Context, params *backends.ListCollectionsParams) (*backends.ListCollectionsResult, error) {
	var result backends.ListCollectionsResult

	exists, err := db.b.metadataStorage.dbExists(db.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !exists {
		return &result, nil
	}

	list, err := db.b.metadataStorage.listCollections(ctx, db.name)
	if err != nil {
		return nil, err
	}

	for _, name := range list {
		result.Collections = append(result.Collections, backends.CollectionInfo{
			Name: name,
		})
	}

	return &result, nil
}

// CreateCollection implements backends.Database interface.
func (db *database) CreateCollection(ctx context.Context, params *backends.CreateCollectionParams) error {
	exists, err := db.b.metadataStorage.dbExists(db.name)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !exists {
		if err = db.create(ctx); err != nil {
			return lazyerrors.Error(err)
		}
	}

	tableName, err := db.b.metadataStorage.createCollection(ctx, db.name, params.Name)
	if err != nil {
		return err
	}

	conn, err := db.b.pool.DB(db.name)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (sjson string)", tableName)

	_, err = conn.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

// DropCollection implements backends.Database interface.
func (db *database) DropCollection(ctx context.Context, params *backends.DropCollectionParams) error {
	table, err := db.b.metadataStorage.tableName(ctx, db.name, params.Name)
	if err != nil {
		return err
	}

	err = db.b.metadataStorage.removeCollection(ctx, db.name, params.Name)
	if err != nil {
		return err
	}

	conn, err := db.b.pool.DB(db.name)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DROP TABLE %s", table)

	_, err = conn.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (db *database) create(ctx context.Context) error {
	f, err := os.Create(filepath.Join(db.b.dir, db.name+dbExtension))
	if err != nil {
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}

	err = db.b.metadataStorage.createDatabase(ctx, db.name)
	if err != nil {
		return err
	}

	return nil
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
