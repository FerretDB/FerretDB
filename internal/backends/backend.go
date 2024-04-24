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
	"cmp"
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/password"
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

// CreateUser stores a new user in the given database and Backend.
func CreateUser(ctx context.Context, b Backend, mechanisms *types.Array, dbName, username, password string) error {
	credentials, err := MakeCredentials(mechanisms, username, password)
	if err != nil {
		return err
	}

	id := uuid.New()
	saved := must.NotFail(types.NewDocument(
		"_id", dbName+"."+username,
		"credentials", credentials,
		"user", username,
		"db", dbName,
		"roles", types.MakeArray(0),
		"userId", types.Binary{Subtype: types.BinaryUUID, B: must.NotFail(id.MarshalBinary())},
	))

	adminDB, err := b.Database("admin")
	if err != nil {
		return err
	}

	users, err := adminDB.Collection("system.users")
	if err != nil {
		return err
	}

	_, err = users.InsertAll(ctx, &InsertAllParams{
		Docs: []*types.Document{saved},
	})
	if err != nil {
		return err
	}

	return nil
}

// MakeCredentials creates a document with credentials for the chosen mechanisms.
func MakeCredentials(mechanisms *types.Array, username, userPassword string) (*types.Document, error) {
	credentials := types.MakeDocument(0)

	// when mechanisms is not specified default is SCRAM-SHA-1
	if mechanisms == nil {
		mechanisms = must.NotFail(types.NewArray("SCRAM-SHA-1"))
	}

	iter := mechanisms.Iterator()
	defer iter.Close()

	for {
		var v any
		_, v, err := iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		var pe error
		var hash *types.Document

		switch v {
		case "PLAIN":
			credentials.Set("PLAIN", must.NotFail(password.PlainHash(userPassword)))
		case "SCRAM-SHA-1":
			hash, pe = password.SCRAMSHA1Hash(username, userPassword)
			if pe != nil {
				return nil, pe
			}

			credentials.Set("SCRAM-SHA-1", hash)
		case "SCRAM-SHA-256":
			hash, pe = password.SCRAMSHA256Hash(userPassword)
			if pe != nil {
				return nil, pe
			}

			credentials.Set("SCRAM-SHA-256", hash)
		default:
			return nil, pe
		}
	}

	return credentials, nil
}

// StatusParams represents the parameters of Backend.Status method.
type StatusParams struct{}

// StatusResult represents the results of Backend.Status method.
type StatusResult struct {
	CountCollections       int64
	CountCappedCollections int32
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
type ListDatabasesParams struct {
	Name string
}

// ListDatabasesResult represents the results of Backend.ListDatabases method.
type ListDatabasesResult struct {
	Databases []DatabaseInfo
}

// DatabaseInfo represents information about a single database.
type DatabaseInfo struct {
	Name string
}

// ListDatabases returns a list of databases sorted by name.
//
// If ListDatabasesParams' Name is not empty, then only the database with that name should be returned (or an empty list).
//
// Database may not exist; that's not an error.
func (bc *backendContract) ListDatabases(ctx context.Context, params *ListDatabasesParams) (*ListDatabasesResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := bc.b.ListDatabases(ctx, params)
	checkError(err)

	if res != nil && len(res.Databases) > 0 {
		must.BeTrue(slices.IsSortedFunc(res.Databases, func(a, b DatabaseInfo) int {
			return cmp.Compare(a.Name, b.Name)
		}))

		if params != nil && params.Name != "" {
			must.BeTrue(len(res.Databases) == 1)
			must.BeTrue(res.Databases[0].Name == params.Name)
		}
	}

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
