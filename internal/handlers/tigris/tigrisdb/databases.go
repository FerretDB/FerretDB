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

// CreateDatabaseIfNotExists ensures that given database exists.
// If the database doesn't exist, it creates it.
// It returns true if the database was created.
func (tdb *TigrisDB) CreateDatabaseIfNotExists(ctx context.Context, db string) (bool, error) {
	exists, err := tdb.databaseExists(ctx, db)
	if err != nil {
		return false, lazyerrors.Error(err)
	}
	if exists {
		return false, nil
	}

	// Database does not exist. Try to create it,
	// but keep in mind that it can be created in concurrent connection.
	err = tdb.driver.CreateDatabase(ctx, db)
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

// databaseExists returns true if database exists.
func (tdb *TigrisDB) databaseExists(ctx context.Context, db string) (bool, error) {
	_, err := tdb.driver.DescribeDatabase(ctx, db)
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
