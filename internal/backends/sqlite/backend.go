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
	"cmp"
	"context"
	"slices"

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
	_   struct{} // prevent unkeyed literals
}

// NewBackend creates a new Backend.
func NewBackend(params *NewBackendParams) (backends.Backend, error) {
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

// Status implements backends.Backend interface.
func (b *backend) Status(ctx context.Context, params *backends.StatusParams) (*backends.StatusResult, error) {
	// since authentication is not supported yet, and there is no way to not establish an SQLite "connection",
	// there is no need to use conninfo
	// TODO https://github.com/FerretDB/FerretDB/issues/3008

	dbs := b.r.DatabaseList(ctx)

	var res backends.StatusResult

	for _, dbName := range dbs {
		cs, err := b.r.CollectionList(ctx, dbName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res.CountCollections += int64(len(cs))

		colls, err := newDatabase(b.r, dbName).ListCollections(ctx, new(backends.ListCollectionsParams))
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		for _, cInfo := range colls.Collections {
			if cInfo.Capped() {
				res.CountCappedCollections++
			}
		}
	}

	return &res, nil
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

	var res *backends.ListDatabasesResult

	if params != nil && len(params.Name) > 0 {
		res = &backends.ListDatabasesResult{
			Databases: make([]backends.DatabaseInfo, 0, 1),
		}
		_, found := slices.BinarySearchFunc(list, params.Name, func(dbName, t string) int {
			return cmp.Compare(dbName, t)
		})
		if found {
			res.Databases = append(res.Databases, backends.DatabaseInfo{
				Name: params.Name,
			})
		}
		return res, nil
	}

	res = &backends.ListDatabasesResult{
		Databases: make([]backends.DatabaseInfo, len(list)),
	}

	for i, dbName := range list {
		res.Databases[i] = backends.DatabaseInfo{
			Name: dbName,
		}
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
