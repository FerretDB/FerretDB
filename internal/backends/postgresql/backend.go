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
	r      *metadata.Registry
	vendor Vendor
	_      struct{} // prevent unkeyed literals
}

// NewBackendParams represents the parameters of NewBackend function.
//
//nolint:vet // for readability
type NewBackendParams struct {
	URI    string
	Vendor Vendor
	L      *zap.Logger
	P      *state.Provider
	_      struct{} // prevent unkeyed literals
}

// NewBackend creates a new backend.
func NewBackend(params *NewBackendParams) (backends.Backend, error) {
	if params.Vendor == 0 {
		panic("unset vendor")
	}

	r, err := metadata.NewRegistry(params.URI, params.L, params.P)
	if err != nil {
		return nil, err
	}

	return backends.BackendContract(&backend{
		r:      r,
		vendor: params.Vendor,
	}), nil
}

// Close implements backends.Backend interface.
func (b *backend) Close() {
	b.r.Close()
}

// Status implements backends.Backend interface.
func (b *backend) Status(ctx context.Context, params *backends.StatusParams) (*backends.StatusResult, error) {
	dbs, err := b.r.DatabaseList(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res backends.StatusResult

	var pingSucceeded bool

	for _, dbName := range dbs {
		var cs []*metadata.Collection

		if cs, err = b.r.CollectionList(ctx, dbName); err != nil {
			return nil, lazyerrors.Error(err)
		}

		res.CountCollections += int64(len(cs))

		colls, err := newDatabase(b.r, b.vendor, dbName).ListCollections(ctx, new(backends.ListCollectionsParams))
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		for _, cInfo := range colls.Collections {
			if cInfo.Capped() {
				res.CountCappedCollections++
			}
		}

		if pingSucceeded {
			continue
		}

		p, err := b.r.DatabaseGetExisting(ctx, dbName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if p == nil {
			continue
		}

		if err = p.Ping(ctx); err != nil {
			return nil, lazyerrors.Error(err)
		}

		pingSucceeded = true
	}

	return &res, nil
}

// Database implements backends.Backend interface.
func (b *backend) Database(name string) (backends.Database, error) {
	return newDatabase(b.r, b.vendor, name), nil
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
		res.Databases[i] = backends.DatabaseInfo{
			Name: dbName,
		}
	}

	return res, nil
}

// DropDatabase implements backends.Backend interface.
func (b *backend) DropDatabase(ctx context.Context, params *backends.DropDatabaseParams) error {
	dropped, err := b.r.DatabaseDrop(ctx, params.Name)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !dropped {
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
