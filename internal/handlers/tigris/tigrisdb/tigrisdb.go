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

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"

	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
)

// TigrisDB represents a Tigris database connection.
type TigrisDB struct {
	driver driver.Driver
	logger *zap.Logger
}

// New returns a new TigrisDB.
func New(cfg *config.Driver) (*TigrisDB, error) {
	d, err := driver.NewDriver(context.TODO(), cfg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return &TigrisDB{
		driver: d,
	}, nil
}

// InTransaction wraps the given function f in a transaction.
// If f returns an error, the transaction is rolled back.
// Errors are wrapped with lazyerrors.Error,
// so the caller needs to use errors.Is to check the error,
// for example, errors.Is(err, *driver.Error).
func (tdb *TigrisDB) InTransaction(ctx context.Context, db string, f func(tx driver.Tx) error) (err error) {
	var tx driver.Tx
	if tx, err = tdb.driver.BeginTx(ctx, db); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	defer func() {
		if err == nil {
			return
		}
		if rerr := tx.Rollback(ctx); rerr != nil {
			tdb.logger.Error("failed to perform rollback", zap.Error(rerr))
		}
	}()

	if err = f(tx); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	if err = tx.Commit(ctx); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	return
}
