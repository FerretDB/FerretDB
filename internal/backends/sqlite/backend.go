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
	uri, err := validateParams(params)
	if err != nil {
		return nil, err
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

func validateParams(params *NewBackendParams) (*url.URL, error) {
	uri, err := url.Parse(params.URI)
	if err != nil {
		return nil, lazyerrors.Errorf("failed to parse backend URI: %w", err)
	}

	if uri.Scheme != "file" {
		return nil, lazyerrors.Errorf("backend URI should have file scheme: %q", params.URI)
	}

	if uri.User != nil {
		return nil, lazyerrors.Errorf("backend URI should not contain user: %q", params.URI)
	}

	if uri.Host != "" {
		return nil, lazyerrors.Errorf("backend URI should not contain host: %q", params.URI)
	}

	dir := uri.Path

	if uri.Opaque != "" {
		dir = uri.Opaque
	}

	if !strings.HasSuffix(dir, "/") {
		return nil, lazyerrors.Errorf("backend URI should be a directory: %q", params.URI)
	}

	fi, err := os.Stat(dir)
	if err != nil {
		return nil, lazyerrors.Errorf("%q should be an existing directory: %w", params.URI, err)
	}

	if !fi.IsDir() {
		return nil, lazyerrors.Errorf("%q should be an existing directory", params.URI)
	}

	uri.Path = dir

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
