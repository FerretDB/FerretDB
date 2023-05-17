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
	list, err := db.b.metadataStorage.ListCollections(ctx, db.name)
	if err != nil {
		return nil, err
	}

	var result backends.ListCollectionsResult

	for _, name := range list {
		result.Collections = append(result.Collections, backends.CollectionInfo{
			Name: name,
		})
	}

	return &result, nil
}

// CreateCollection implements backends.Database interface.
func (db *database) CreateCollection(ctx context.Context, params *backends.CreateCollectionParams) error {
	_, err := db.b.metadataStorage.CreateCollection(ctx, db.name, params.Name)
	if err != nil {
		return err
	}

	conn, err := db.b.pool.DB(db.name)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (sjson string)", params.Name)

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

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
