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

package tigrisdb

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// createDatabaseIfNotExists ensures that given database exists.
// If the database doesn't exist, it creates it.
// It returns true if the database was created.
func (tdb *TigrisDB) createDatabaseIfNotExists(ctx context.Context, db string) (bool, error) {
	exists, err := tdb.databaseExists(ctx, db)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	if exists {
		return false, nil
	}

	// Database does not exist. Try to create it,
	// but keep in mind that it can be created in concurrent connection.
	// If we detect that other creation is in flight, we give up to retriesMax attempts to create the database.
	// TODO https://github.com/FerretDB/FerretDB/issues/1341
	// TODO https://github.com/FerretDB/FerretDB/issues/1720
	const (
		retriesMax    = 3
		retryDelayMin = 100 * time.Millisecond
		retryDelayMax = 200 * time.Millisecond
	)

	for i := 0; i < retriesMax; i++ {
		_, err = tdb.Driver.CreateProject(ctx, db)
		tdb.l.Debug("createDatabaseIfNotExists", zap.String("db", db), zap.Error(err))

		var driverErr *driver.Error

		switch {
		case err == nil:
			return true, nil
		case errors.As(err, &driverErr):
			if IsAlreadyExists(err) {
				return false, nil
			}

			if isOtherCreationInFlight(err) {
				deltaMS := rand.Int63n((retryDelayMax - retryDelayMin).Milliseconds())
				ctxutil.Sleep(ctx, retryDelayMin+time.Duration(deltaMS)*time.Millisecond)
				continue
			}

			return false, lazyerrors.Error(err)
		default:
			return false, lazyerrors.Error(err)
		}
	}

	return false, lazyerrors.Error(err)
}

// databaseExists returns true if database exists.
func (tdb *TigrisDB) databaseExists(ctx context.Context, db string) (bool, error) {
	_, err := tdb.Driver.DescribeDatabase(ctx, db)
	switch err := err.(type) {
	case nil:
		return true, nil
	case *driver.Error:
		if IsNotFound(err) {
			return false, nil
		}

		return false, lazyerrors.Error(err)
	default:
		return false, lazyerrors.Error(err)
	}
}
