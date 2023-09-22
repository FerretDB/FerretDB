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

package backends

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// Backend is a generic interface for all backends for accessing them.
//
// Backend object should be stateful and wrap database connection(s).
// Handler uses only one long-lived Backend object.
//
// Backend(s) methods can be called by multiple client connections / command handlers concurrently.
// They should be thread-safe.
//
// See backendContract and its methods for additional details.
type Backend interface {
	Close()

	Name() string
	Status(context.Context, *StatusParams) (*StatusResult, error)

	Database(string) (Database, error)
	ListDatabases(context.Context, *ListDatabasesParams) (*ListDatabasesResult, error)
	DropDatabase(context.Context, *DropDatabaseParams) error

	prometheus.Collector

	// There is no interface method to create a database; see package documentation.
}

// backendContract implements Backend interface.
type backendContract struct {
	b     Backend
	token *resource.Token
}

// BackendContract wraps Backend and enforces its contract.
//
// All backend implementations should use that function when they create new Backend instances.
// The handler should not use that function.
//
// See backendContract and its methods for additional details.
func BackendContract(b Backend) Backend {
	bc := &backendContract{
		b:     b,
		token: resource.NewToken(),
	}
	resource.Track(bc, bc.token)

	return bc
}

// Close closes all database connections and frees all resources associated with the backend.
func (bc *backendContract) Close() {
	bc.b.Close()

	resource.Untrack(bc, bc.token)
}

// Name returns human-readable formatted name of the backend.
func (bc *backendContract) Name() string {
	return bc.b.Name()
}

// StatusParams represents the parameters of Backend.Status method.
type StatusParams struct{}

// StatusResult represents the results of Backend.Status method.
type StatusResult struct {
	CountCollections int64
}

// Status returns backend's status.
//
// This method should also be used to check that the backend is alive,
// connection can be established and authenticated.
// For that reason, the implementation should not return only cached results.
func (bc *backendContract) Status(ctx context.Context, params *StatusParams) (*StatusResult, error) {
	defer observability.FuncCall(ctx)()

	// to both check that conninfo is present (which is important for that method),
	// and to render doc.go correctly
	must.NotBeZero(conninfo.Get(ctx))

	res, err := bc.b.Status(ctx, params)
	checkError(err)

	return res, err
}

// Database returns a Database instance for the given valid name.
//
// The database does not need to exist.
func (bc *backendContract) Database(name string) (Database, error) {
	var res Database

	err := validateDatabaseName(name)
	if err == nil {
		res, err = bc.b.Database(name)
	}

	checkError(err, ErrorCodeDatabaseNameIsInvalid)

	return res, err
}

// ListDatabasesParams represents the parameters of Backend.ListDatabases method.
type ListDatabasesParams struct{}

// ListDatabasesResult represents the results of Backend.ListDatabases method.
type ListDatabasesResult struct {
	Databases []DatabaseInfo
}

// DatabaseInfo represents information about a single database.
type DatabaseInfo struct {
	Name string
	Size int64
}

// ListDatabases returns a list of databases.
func (bc *backendContract) ListDatabases(ctx context.Context, params *ListDatabasesParams) (*ListDatabasesResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := bc.b.ListDatabases(ctx, params)
	checkError(err)

	return res, err
}

// DropDatabaseParams represents the parameters of Backend.DropDatabase method.
type DropDatabaseParams struct {
	Name string
}

// DropDatabase drops existing database for given parameters (including valid name).
func (bc *backendContract) DropDatabase(ctx context.Context, params *DropDatabaseParams) error {
	defer observability.FuncCall(ctx)()

	err := validateDatabaseName(params.Name)
	if err == nil {
		err = bc.b.DropDatabase(ctx, params)
	}

	checkError(err, ErrorCodeDatabaseNameIsInvalid, ErrorCodeDatabaseDoesNotExist)

	return err
}

// Describe implements prometheus.Collector.
func (bc *backendContract) Describe(ch chan<- *prometheus.Desc) {
	bc.b.Describe(ch)
}

// Collect implements prometheus.Collector.
func (bc *backendContract) Collect(ch chan<- prometheus.Metric) {
	bc.b.Collect(ch)
}

// check interfaces
var (
	_ Backend = (*backendContract)(nil)
)
