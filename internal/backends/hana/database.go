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

package hana

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// database implements backends.Database.
type database struct{}

// Collection implements backends.Database interface.
func (db *database) Collection(name string) (backends.Collection, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// ListCollections implements backends.Database interface.
//
//nolint:lll // for readability
func (db *database) ListCollections(ctx context.Context, params *backends.ListCollectionsParams) (*backends.ListCollectionsResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// CreateCollection implements backends.Database interface.
func (db *database) CreateCollection(ctx context.Context, params *backends.CreateCollectionParams) error {
	return lazyerrors.New("not implemented yet")
}

// DropCollection implements backends.Database interface.
func (db *database) DropCollection(ctx context.Context, params *backends.DropCollectionParams) error {
	return lazyerrors.New("not implemented yet")
}

// RenameCollection implements backends.Database interface.
func (db *database) RenameCollection(ctx context.Context, params *backends.RenameCollectionParams) error {
	return lazyerrors.New("not implemented yet")
}

// Stats implements backends.Database interface.
func (db *database) Stats(ctx context.Context, params *backends.DatabaseStatsParams) (*backends.DatabaseStatsResult, error) {
	return nil, lazyerrors.New("not implemented yet")
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
