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

	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// Backend is a generic interface for all backends for accessing them.
//
// Backend object is expected to be stateful and wrap database connection(s).
// Handler can create one Backend or multiple Backends with different authentication credentials.
//
// Backend(s) methods can be called by multiple client connections / command handlers concurrently.
// They should be thread-safe.
//
// See backendContract and its methods for additional details.
type Backend interface {
	Close()
	Database(string) (Database, error)
	ListDatabases(context.Context, *ListDatabasesParams) (*ListDatabasesResult, error)
	DropDatabase(context.Context, *DropDatabaseParams) error

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

// Database returns a Database instance for the given name.
//
// The database does not need to exist; even parameters like name could be invalid FIXME.
func (bc *backendContract) Database(name string) (Database, error) {
	var res Database
	err := validDatabaseName(name)
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

// ListDatabases returns a Database instance for given parameters.
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

// DropDatabase drops existing database for given parameters.
func (bc *backendContract) DropDatabase(ctx context.Context, params *DropDatabaseParams) error {
	defer observability.FuncCall(ctx)()

	err := validDatabaseName(params.Name)
	if err == nil {
		err = bc.b.DropDatabase(ctx, params)
	}

	checkError(err, ErrorCodeDatabaseNameIsInvalid, ErrorCodeDatabaseDoesNotExist)

	return err
}

// check interfaces
var (
	_ Backend = (*backendContract)(nil)
)
