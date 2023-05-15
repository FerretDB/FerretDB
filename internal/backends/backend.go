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

// Backends is a generic interface for all backends for accessing them.
//
// Backend object is expected to be stateful and wrap database connection(s).
// Handler can create one Backend or multiple Backends with different authentication credentials.
//
// Backend(s) methods can be called by multiple client connections / command handlers concurrently.
// They should be thread-safe.
//
// See backendContract and its methods for additional details.
type Backend interface {
	Database(*DatabaseParams) Database
	ListDatabases(*ListDatabasesParams) (*ListDatabasesResult, error)
}

// BackendContract wraps Backend and enforces its contract.
//
// All backend implementations should use that function when they create new Backend instances.
// The handler should not use that function.
//
// See backendContract and its methods for additional details.
func BackendContract(b Backend) Backend {
	return &backendContract{
		b: b,
	}
}

// backendContract implements Backend interface.
type backendContract struct {
	b Backend
}

// DatabaseParams represents the parameters of Backend.Database method.
type DatabaseParams struct {
	Name string
}

// Database returns a Database instance for given parameters.
//
// The database does not need to exist; even parameters like name could be invalid.
func (bc *backendContract) Database(params *DatabaseParams) Database {
	return bc.b.Database(params)
}

type ListDatabasesParams struct{}

type ListDatabasesResult struct{}

func (bc *backendContract) ListDatabases(params *ListDatabasesParams) (res *ListDatabasesResult, err error) {
	defer checkError(err)
	res, err = bc.b.ListDatabases(params)
	return
}

// check interfaces
var (
	_ Backend = (*backendContract)(nil)
)
