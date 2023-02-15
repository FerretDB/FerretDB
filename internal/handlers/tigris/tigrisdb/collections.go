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

	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// CreateCollectionIfNotExist ensures that given collection exist.
// If needed, it creates both database and collection.
// It returns true if the collection was created.
func (tdb *TigrisDB) CreateCollectionIfNotExist(ctx context.Context, db, collection string, schema driver.Schema) (bool, error) {
	_, err := tdb.createDatabaseIfNotExists(ctx, db)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	exists, err := tdb.CollectionExists(ctx, db, collection)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	if exists {
		return false, nil
	}

	err = tdb.Driver.UseDatabase(db).CreateOrUpdateCollection(ctx, collection, schema)
	tdb.l.Debug(
		"CreateCollectionIfNotExist",
		zap.String("db", db), zap.String("collection", collection), zap.ByteString("schema", schema), zap.Error(err),
	)

	var driverErr *driver.Error

	switch {
	case err == nil:
		return true, nil
	case errors.As(err, &driverErr):
		if IsAlreadyExists(err) {
			return false, nil
		}

		return false, err
	default:
		return false, lazyerrors.Error(err)
	}
}

// CollectionExists returns true if collection exists.
func (tdb *TigrisDB) CollectionExists(ctx context.Context, db, collection string) (bool, error) {
	_, err := tdb.Driver.UseDatabase(db).DescribeCollection(ctx, collection)
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
