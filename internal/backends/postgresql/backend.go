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
)

// backend implements backends.Backend interface.
type backend struct {
	version string
}

// NewBackendParams represents the parameters of NewBackend function.
//
//nolint:vet // for readability
type NewBackendParams struct {
	URI     string
	L       *zap.Logger
	Version string
}

// NewBackend creates a new backend for PostgreSQL-compatible database.
func NewBackend(params *NewBackendParams) (backends.Backend, error) {
	return backends.BackendContract(&backend{
		version: params.Version,
	}), nil
}

// Close implements backends.Backend interface.
func (b *backend) Close() {
}

// Database implements backends.Backend interface.
func (b *backend) Database(name string) (backends.Database, error) {
	return newDatabase(name), nil
}

// ListDatabases implements backends.Backend interface.
//
//nolint:lll // for readability
func (b *backend) ListDatabases(ctx context.Context, params *backends.ListDatabasesParams) (*backends.ListDatabasesResult, error) {
	panic("not implemented")
}

// DropDatabase implements backends.Backend interface.
func (b *backend) DropDatabase(ctx context.Context, params *backends.DropDatabaseParams) error {
	panic("not implemented")
}

// Info implements backends.Backend interface.
func (b *backend) Info() *backends.Info {
	return &backends.Info{
		Name:    "PostgreSQL",
		Version: b.version,
	}
}

// Describe implements prometheus.Collector.
func (b *backend) Describe(ch chan<- *prometheus.Desc) {
	panic("not implemented")
}

// Collect implements prometheus.Collector.
func (b *backend) Collect(ch chan<- prometheus.Metric) {
	panic("not implemented")
}

// check interfaces
var (
	_ backends.Backend = (*backend)(nil)
)
