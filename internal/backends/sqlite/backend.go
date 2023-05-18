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
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/FerretDB/FerretDB/internal/backends"
)

// backend implements backends.Backend interface.
type backend struct {
	dir             string
	pool            *connPool
	metadataStorage *metadataStorage
}

// NewBackendParams represents the parameters of NewBackend function.
type NewBackendParams struct {
	Dir string
}

// NewBackend creates a new SQLite backend.
func NewBackend(params *NewBackendParams) (backends.Backend, error) {
	pool := newConnPool()

	storage, err := newMetadataStorage(params.Dir, pool)
	if err != nil {
		return nil, err
	}

	return backends.BackendContract(&backend{
		dir:             params.Dir,
		pool:            pool,
		metadataStorage: storage,
	}), nil
}

// Database implements backends.Backend interface.
func (b *backend) Database(name string) backends.Database {
	return newDatabase(b, name)
}

// ListDatabases implements backends.Backend interface.
//
//nolint:lll // for readability
func (b *backend) ListDatabases(ctx context.Context, params *backends.ListDatabasesParams) (*backends.ListDatabasesResult, error) {
	list, err := b.metadataStorage.listDatabases()
	if err != nil {
		return nil, err
	}

	var result backends.ListDatabasesResult
	for _, db := range list {
		result.Databases = append(result.Databases, backends.DatabaseInfo{Name: db})
	}

	return &result, nil
}

// DropDatabase implements backends.Backend interface.
func (b *backend) DropDatabase(ctx context.Context, params *backends.DropDatabaseParams) error {
	err := os.Remove(filepath.Join(b.dir, params.Name+dbExtension))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Close implements backends.Backend interface.
func (b *backend) Close() error {
	return b.pool.Close()
}

// check interfaces
var (
	_ backends.Backend = (*backend)(nil)
)
