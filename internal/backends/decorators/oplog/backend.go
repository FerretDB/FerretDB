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

// Package oplog provides decorators that add OpLog functionality to the backend.
package oplog

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
)

// backend implements backends.Backend interface by delegating all methods to the wrapped backend.
type backend struct {
	b backends.Backend
	l *zap.Logger
}

// NewBackend creates a new backend that wraps the given backend.
func NewBackend(b backends.Backend, l *zap.Logger) backends.Backend {
	return &backend{
		b: b,
		l: l,
	}
}

// Close implements backends.Backend interface.
func (b *backend) Close() {
	b.b.Close()
}

// Database implements backends.Backend interface.
func (b *backend) Database(name string) (backends.Database, error) {
	db, err := b.b.Database(name)
	if err != nil {
		return nil, err
	}

	return newDatabase(db, b.l), nil
}

// ListDatabases implements backends.Backend interface.
//
//nolint:lll // for readability
func (b *backend) ListDatabases(ctx context.Context, params *backends.ListDatabasesParams) (*backends.ListDatabasesResult, error) {
	return b.b.ListDatabases(ctx, params)
}

// DropDatabase implements backends.Backend interface.
func (b *backend) DropDatabase(ctx context.Context, params *backends.DropDatabaseParams) error {
	return b.b.DropDatabase(ctx, params)
}

// Name implements backends.Backend interface.
func (b *backend) Name() string {
	return b.b.Name()
}

// Describe implements prometheus.Collector.
func (b *backend) Describe(ch chan<- *prometheus.Desc) {
	b.b.Describe(ch)
}

// Collect implements prometheus.Collector.
func (b *backend) Collect(ch chan<- prometheus.Metric) {
	b.b.Collect(ch)
}

// check interfaces
var (
	_ backends.Backend = (*backend)(nil)
)
