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

// Package tigrisdb provides Tigris connection utilities.
package tigrisdb

import (
	"context"

	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// TigrisDB represents concurrency-safe TigrisDB connection pool.
type TigrisDB struct {
	Driver driver.Driver
	l      *zap.Logger
}

// New returns a new TigrisDB connection pool.
//
// Passed ctx is used only by the first checking connection.
// Canceling it after that function returns does nothing.
func New(ctx context.Context, cfg *config.Driver, logger *zap.Logger) (*TigrisDB, error) {
	d, err := driver.NewDriver(ctx, cfg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, err = d.Health(ctx); err != nil {
		_ = d.Close()
		return nil, err
	}

	return &TigrisDB{
		Driver: d,
		l:      logger,
	}, nil
}
