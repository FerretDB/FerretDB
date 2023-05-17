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
	"os"
	"path/filepath"

	"github.com/FerretDB/FerretDB/internal/backends"
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

// Collection implements backends.Database interface.
func (db *database) Collection(name string) backends.Collection {
	return newCollection(db, name)
}

// ListCollections implements backends.Database interface.
//
//nolint:lll // for readability
func (db *database) ListCollections(ctx context.Context, params *backends.ListCollectionsParams) (*backends.ListCollectionsResult, error) {
	var result backends.ListCollectionsResult

	list, err := db.b.metadataStorage.ListCollections(ctx, db.name)
	if errors.Is(err, errDatabaseNotFound) {
		if err = db.create(ctx); err != nil {
			return &result, err
		}
	}
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
	tableName, err := db.b.metadataStorage.CreateCollection(ctx, db.name, params.Name)
	if errors.Is(err, errDatabaseNotFound) {
		if err = db.create(ctx); err != nil {
			return err
		}
	}
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
	table, err := db.b.metadataStorage.CollectionInfo(db.name, params.Name)
	if err != nil {
		return err
	}

	err = db.b.metadataStorage.RemoveCollection(db.name, params.Name)
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
	_, err := os.Create(filepath.Join(db.b.dir, db.name+dbExtension))
	if err != nil {
		return err
	}

	db.b.metadataStorage.CreateDatabase(nil, db.name)
	return nil
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
