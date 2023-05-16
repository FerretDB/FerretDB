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

	_ "modernc.org/sqlite"

	"github.com/FerretDB/FerretDB/internal/backends"
)

type backend struct{}

// NewBackend creates a new SQLite backend.
func NewBackend() backends.Backend {
	return backends.BackendContract(new(backend))
}

// Database implements backends.Backend interface.
func (b *backend) Database(ctx context.Context, params *backends.DatabaseParams) backends.Database {
	return newDatabase(b)
}

// ListDatabases implements backends.Backend interface.
//
//nolint:lll // for readability
func (b *backend) ListDatabases(ctx context.Context, params *backends.ListDatabasesParams) (*backends.ListDatabasesResult, error) {
	panic("not implemented") // TODO: Implement
}

// DropDatabase implements backends.Backend interface.
func (b *backend) DropDatabase(ctx context.Context, params *backends.DropDatabaseParams) error {
	panic("not implemented") // TODO: Implement
}

// check interfaces
var (
	_ backends.Backend = (*backend)(nil)
)
