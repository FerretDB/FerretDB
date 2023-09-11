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
)

// database implements backends.Database interface.
type database struct {
	name string
}

// newDatabase creates a new Database.
func newDatabase(name string) backends.Database {
	return backends.DatabaseContract(&database{
		name: name,
	})
}

// Collection implements backends.Database interface.
func (db *database) Collection(name string) (backends.Collection, error) {
	return newCollection(db.name, name), nil
}

// ListCollections implements backends.Database interface.
//
//nolint:lll // for readability
func (db *database) ListCollections(ctx context.Context, params *backends.ListCollectionsParams) (*backends.ListCollectionsResult, error) {
	panic("not implemented")
}

// CreateCollection implements backends.Database interface.
func (db *database) CreateCollection(ctx context.Context, params *backends.CreateCollectionParams) error {
	panic("not implemented")
}

// DropCollection implements backends.Database interface.
func (db *database) DropCollection(ctx context.Context, params *backends.DropCollectionParams) error {
	panic("not implemented")
}

// RenameCollection implements backends.Database interface.
func (db *database) RenameCollection(ctx context.Context, params *backends.RenameCollectionParams) error {
	panic("not implemented")
}

// Stats implements backends.Database interface.
func (db *database) Stats(ctx context.Context, params *backends.DatabaseStatsParams) (*backends.DatabaseStatsResult, error) {
	panic("not implemented")
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
