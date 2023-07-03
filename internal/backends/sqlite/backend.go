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
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// backend implements backends.Backend interface.
type backend struct {
	r   *metadata.Registry
	uri *url.URL
}

// NewBackendParams represents the parameters of NewBackend function.
type NewBackendParams struct {
	URI string
	L   *zap.Logger
}

// NewBackend creates a new SQLite backend.
func NewBackend(params *NewBackendParams) (backends.Backend, error) {
	uri, err := validateURI(params.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQLite URI %q: %s", params.URI, err)
	}

	r, err := metadata.NewRegistry(uri, params.L.Named("metadata"))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return backends.BackendContract(&backend{
		r:   r,
		uri: uri,
	}), nil
}

// validateURI checks given URI value and returns parsed URL.
// URI should contain 'file' scheme and point to an existing directory.
// Path should end with '/'. Authority should be empty or absent.
//
// Returned URL contains path in both Path and Opaque to make String() method work correctly.
func validateURI(value string) (*url.URL, error) {
	uri, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	if uri.Scheme != "file" {
		return nil, fmt.Errorf(`expected "file:" schema, got %q`, uri.Scheme)
	}

	if uri.User != nil {
		return nil, fmt.Errorf(`expected empty user info, got %q`, uri.User)
	}

	if uri.Host != "" {
		return nil, fmt.Errorf(`expected empty host, got %q`, uri.Host)
	}

	if uri.Path == "" && uri.Opaque != "" {
		uri.Path = uri.Opaque
	}
	uri.Opaque = uri.Path

	if !strings.HasSuffix(uri.Path, "/") {
		return nil, fmt.Errorf(`expected path ending with "/", got %q`, uri.Host)
	}

	fi, err := os.Stat(uri.Path)
	if err != nil {
		return nil, fmt.Errorf(`%q should be an existing directory, got %s`, uri.Path, err)
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf(`%q should be an existing directory`, uri.Path)
	}

	return uri, nil
}

// Close implements backends.Backend interface.
func (b *backend) Close() {
	b.r.Close()
}

// Database implements backends.Backend interface.
func (b *backend) Database(name string) backends.Database {
	return newDatabase(b.r, name)
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

// check interfaces
var (
	_ backends.Backend = (*backend)(nil)
)
