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

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
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

// NewBackend creates a new backend for PostgreSQL-compatible database.
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
}

// Name implements backends.Backend interface.
func (b *backend) Name() string {
	return "PostgreSQL"
}

// Status implements backends.Backend interface.
func (b *backend) Status(ctx context.Context, params *backends.StatusParams) (*backends.StatusResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3404
	return new(backends.StatusResult), nil
}

// Database implements backends.Backend interface.
func (b *backend) Database(name string) (backends.Database, error) {
	return newDatabase(b.r, name), nil
}

// ListDatabases implements backends.Backend interface.
//
//nolint:lll // for readability
func (b *backend) ListDatabases(ctx context.Context, params *backends.ListDatabasesParams) (*backends.ListDatabasesResult, error) {
	list, err := b.r.DatabaseList(ctx)
	if err != nil {
		return nil, err
	}

	res := &backends.ListDatabasesResult{
		Databases: make([]backends.DatabaseInfo, len(list)),
	}

	for i, dbName := range list {
		db, err := b.Database(dbName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		stats, err := db.Stats(ctx, new(backends.DatabaseStatsParams))
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseDoesNotExist) {
			stats = new(backends.DatabaseStatsResult)
			err = nil
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res.Databases[i] = backends.DatabaseInfo{
			Name: dbName,
			Size: stats.SizeTotal,
		}
	}

	return res, nil
}

// DropDatabase implements backends.Backend interface.
func (b *backend) DropDatabase(ctx context.Context, params *backends.DropDatabaseParams) error {
	// TODO https://github.com/FerretDB/FerretDB/issues/3404
	return nil
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
