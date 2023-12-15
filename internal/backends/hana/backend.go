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
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// backend implements backends.Backend interface.
type backend struct {
	hdb *fsql.DB
	l   *zap.Logger
}

// NewBackendParams represents the parameters of NewBackend function.
//
//nolint:vet // for readability
type NewBackendParams struct {
	URI string
	L   *zap.Logger
	P   *state.Provider
}

// NewBackend creates a new Backend.
func NewBackend(params *NewBackendParams) (backends.Backend, error) {
	db, err := sql.Open("hdb", params.URI)
	if err != nil {
		return nil, err
	}

	hdb := fsql.WrapDB(db, "hana", params.L)

	return backends.BackendContract(&backend{
		hdb: hdb,
		l:   params.L,
	}), nil
}

// Close implements backends.Backend interface.
func (b *backend) Close() {
	if b.hdb.Close() != nil {
		panic("could not close hana db connection")
	}
}

// Status implements backends.Backend interface.
func (b *backend) Status(ctx context.Context, params *backends.StatusParams) (*backends.StatusResult, error) {
	sqlStmt := "SELECT DISTINCT(SCHEMA_NAME) FROM M_TABLES WHERE TABLE_TYPE = 'COLLECTION'"

	// List out all schemas
	rows, err := b.hdb.QueryContext(ctx, sqlStmt)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	var res backends.StatusResult

	for rows.Next() {
		// db name is schema name from here on out
		var dbName string
		if err = rows.Scan(&dbName); err != nil {
			return nil, lazyerrors.Error(err)
		}

		list, errDB := newDatabase(b.hdb, dbName).ListCollections(ctx, new(backends.ListCollectionsParams))
		if errDB != nil {
			return nil, lazyerrors.Error(err)
		}

		res.CountCollections += int64(len(list.Collections))

		for _, cInfo := range list.Collections {
			if cInfo.Capped() {
				res.CountCappedCollections++
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &res, nil
}

// Database implements backends.Backend interface.
func (b *backend) Database(name string) (backends.Database, error) {
	return newDatabase(b.hdb, name), nil
}

// ListDatabases implements backends.Backend interface.
//
//nolint:lll // for readability
func (b *backend) ListDatabases(ctx context.Context, params *backends.ListDatabasesParams) (*backends.ListDatabasesResult, error) {
	rows, err := b.hdb.QueryContext(ctx, "SELECT SCHEMA_NAME FROM SCHEMAS")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []backends.DatabaseInfo

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		databases = append(databases, backends.DatabaseInfo{Name: dbName})
	}

	return &backends.ListDatabasesResult{Databases: databases}, nil
}

// DropDatabase implements backends.Backend interface.
func (b *backend) DropDatabase(ctx context.Context, params *backends.DropDatabaseParams) error {
	dropped, err := dropSchema(ctx, b.hdb, params.Name)
	if err != nil {
		return getHanaErrorIfExists(err)
	}

	if !dropped {
		return backends.NewError(backends.ErrorCodeDatabaseDoesNotExist, err)
	}

	return nil
}

// Describe implements prometheus.Collector.
func (b *backend) Describe(ch chan<- *prometheus.Desc) {
}

// Collect implements prometheus.Collector.
func (b *backend) Collect(ch chan<- prometheus.Metric) {
}

// check interfaces
var (
	_ backends.Backend = (*backend)(nil)
)
