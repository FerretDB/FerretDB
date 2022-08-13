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

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// CreateCollectionIfNotExists ensures that given database and collection exist.
// If needed, it creates both database and collection.
// It returns true if the collection was created.
func (tdb *TigrisDB) CreateCollectionIfNotExists(ctx context.Context, db, collection string, schema driver.Schema) (bool, error) {
	exists, err := tdb.collectionExists(ctx, db, collection)
	if err != nil {
		return false, lazyerrors.Error(err)
	}
	if exists {
		return false, nil
	}

	// Collection (or even database) does not exist. Try to create it,
	// but keep in mind that it can be created in concurrent connection.
	_, err = tdb.CreateDatabaseIfNotExists(ctx, db)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	err = tdb.driver.UseDatabase(db).CreateOrUpdateCollection(ctx, collection, schema)
	switch err := err.(type) {
	case nil:
		return true, nil
	case *driver.Error:
		if IsAlreadyExists(err) {
			return false, nil
		}
		return false, lazyerrors.Error(err)
	default:
		return false, lazyerrors.Error(err)
	}
}

// collectionExists returns true if collection exists.
func (tdb *TigrisDB) collectionExists(ctx context.Context, db, collection string) (bool, error) {
	_, err := tdb.driver.UseDatabase(db).DescribeCollection(ctx, collection)
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
