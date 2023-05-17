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
	panic("not implemented") // TODO: Implement
}

// DropCollection implements backends.Database interface.
func (db *database) DropCollection(ctx context.Context, params *backends.DropCollectionParams) error {
	panic("not implemented") // TODO: Implement
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
