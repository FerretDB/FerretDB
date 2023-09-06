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

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// backend implements backends.Backend interface.
type backend struct {
	r *metadata.Registry
}

// NewBackendParams represents the parameters of NewBackend function.
//
//nolint:vet // for readability
type NewBackendParams struct {
	URI string
	L   *zap.Logger
	P   *state.Provider
}

// NewBackend creates a new SQLite backend.
func NewBackend(params *NewBackendParams) (backends.Backend, error) {
	if params.P == nil {
		panic("state provider is required but not set")
	}

	r, err := metadata.NewRegistry(params.URI, params.L, params.P)
	if err != nil {
		return nil, err
	}

	return backends.BackendContract(&backend{
		r: r,
	}), nil
}

// Close implements backends.Backend interface.
func (b *backend) Close() {
	b.r.Close()
}

// Database implements backends.Backend interface.
func (b *backend) Database(name string) (backends.Database, error) {
	return newDatabase(b.r, name), nil
}

// ListDatabases implements backends.Backend interface.
//
//nolint:lll // for readability
func (b *backend) ListDatabases(ctx context.Context, params *backends.ListDatabasesParams) (*backends.ListDatabasesResult, error) {
	list := b.r.DatabaseList(ctx)

	res := &backends.ListDatabasesResult{
		Databases: make([]backends.DatabaseInfo, len(list)),
	}
	for i, db := range list {
		res.Databases[i] = backends.DatabaseInfo{Name: db}
	}

	return res, nil
}

// DropDatabase implements backends.Backend interface.
func (b *backend) DropDatabase(ctx context.Context, params *backends.DropDatabaseParams) error {
	if dropped := b.r.DatabaseDrop(ctx, params.Name); !dropped {
		return backends.NewError(backends.ErrorCodeDatabaseDoesNotExist, nil)
	}

	return nil
}

// ServerStatus implements backends.Backend interface.
func (b *backend) ServerStatus(ctx context.Context, params *backends.ServerStatusParams) (*backends.ServerStatusResult, error) {
	dbs := b.r.DatabaseList(ctx)

	res := new(backends.ServerStatusResult)

	for _, dbName := range dbs {
		cs, err := b.r.CollectionList(ctx, dbName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res.CountCollections += int64(len(cs))
	}

	return res, nil
}

// Name implements backends.Backend interface.
func (b *backend) Name() string {
	return "SQLite"
}

// Describe implements prometheus.Collector.
func (b *backend) Describe(ch chan<- *prometheus.Desc) {
	b.r.Describe(ch)
}

// Collect implements prometheus.Collector.
func (b *backend) Collect(ch chan<- prometheus.Metric) {
	b.r.Collect(ch)
}

// check interfaces
var (
	_ backends.Backend = (*backend)(nil)
)
