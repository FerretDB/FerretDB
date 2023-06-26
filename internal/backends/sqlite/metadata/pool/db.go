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

package pool

import (
	"database/sql"

	_ "modernc.org/sqlite" // register database/sql driver

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// db wraps *sql.DB with resource tracking.
type db struct {
	sqlDB *sql.DB
	token *resource.Token
}

// openDB opens existing database.
func openDB(uri string) (*db, error) {
	sqlDB, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/2909

	// TODO https://github.com/FerretDB/FerretDB/issues/2755
	sqlDB.SetConnMaxIdleTime(0)
	sqlDB.SetConnMaxLifetime(0)
	// sqlDB.SetMaxIdleConns(5)
	// sqlDB.SetMaxOpenConns(5)

	if err = sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, lazyerrors.Error(err)
	}

	res := &db{
		sqlDB: sqlDB,
		token: resource.NewToken(),
	}

	resource.Track(res, res.token)

	return res, nil
}

// Close closes db.
func (db *db) Close() error {
	err := db.sqlDB.Close()

	resource.Untrack(db, db.token)

	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
