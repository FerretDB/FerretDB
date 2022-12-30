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

// TigrisDB represents a Tigris database connection.
type TigrisDB struct {
	Driver driver.Driver
	l      *zap.Logger
}

// New returns a new TigrisDB.
//
// Passed context is used only by the first checking connection.
// Canceling it after that function returns does nothing.
//
// If lazy is true, then connectivity is not checked.
// Lazy connections are used by FerretDB when it starts earlier than backend.
// Non-lazy connections are used by tests.
func New(ctx context.Context, cfg *config.Driver, logger *zap.Logger, lazy bool) (*TigrisDB, error) {
	d, err := driver.NewDriver(ctx, cfg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !lazy {
		if _, err = d.Health(ctx); err != nil {
			_ = d.Close()
			return nil, err
		}
	}

	return &TigrisDB{
		Driver: d,
		l:      logger,
	}, nil
}
