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

package sqlitedb

import (
	"database/sql"
	"fmt"
)

func CreateCollection(db, collection string) error {
	database, err := createDatabase(db)
	if err != nil {
		return err
	}

	sqlExpr := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (json string)", collection)

	_, err = database.Exec(sqlExpr)
	if err != nil {
		return err
	}

	return nil
}

func CreateCollectionIfNotExists(db, collection string) (*sql.DB, error) {
	database, err := createDatabase(db)
	if err != nil {
		return nil, err
	}

	sqlExpr := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (json string)", collection)

	_, err = database.Exec(sqlExpr)
	if err != nil {
		return nil, err
	}

	return database, nil
}
