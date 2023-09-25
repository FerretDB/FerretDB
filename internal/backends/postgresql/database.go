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

package postgresql

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
)

// database implements backends.Database interface.
type database struct {
	r    *metadata.Registry
	name string
}

// newDatabase creates a new Database.
func newDatabase(r *metadata.Registry, name string) backends.Database {
	return backends.DatabaseContract(&database{
		r:    r,
		name: name,
	})
}

// Collection implements backends.Database interface.
func (db *database) Collection(name string) (backends.Collection, error) {
	return newCollection(db.r, db.name, name), nil
}

// ListCollections implements backends.Database interface.
//
//nolint:lll // for readability
func (db *database) ListCollections(ctx context.Context, params *backends.ListCollectionsParams) (*backends.ListCollectionsResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3406
	return new(backends.ListCollectionsResult), nil
}

// CreateCollection implements backends.Database interface.
func (db *database) CreateCollection(ctx context.Context, params *backends.CreateCollectionParams) error {
	// TODO https://github.com/FerretDB/FerretDB/issues/3406
	return nil
}

// DropCollection implements backends.Database interface.
func (db *database) DropCollection(ctx context.Context, params *backends.DropCollectionParams) error {
	// TODO https://github.com/FerretDB/FerretDB/issues/3406
	return nil
}

// RenameCollection implements backends.Database interface.
func (db *database) RenameCollection(ctx context.Context, params *backends.RenameCollectionParams) error {
	// TODO https://github.com/FerretDB/FerretDB/issues/3393
	return nil
}

// Stats implements backends.Database interface.
func (db *database) Stats(ctx context.Context, params *backends.DatabaseStatsParams) (*backends.DatabaseStatsResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3402
	return new(backends.DatabaseStatsResult), nil
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
