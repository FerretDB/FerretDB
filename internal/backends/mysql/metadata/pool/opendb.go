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
	"context"
	"database/sql"

	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// openDB opens existing database connections or creates a new one.
func openDB(uri string, l *zap.Logger, sp *state.Provider) (*fsql.DB, error) {
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(100)

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, lazyerrors.Error(err)
	}

	// set backend version
	if sp.Get().BackendVersion == "" {
		err := sp.Update(func(s *state.State) {
			s.BackendName = "MySQL"

			row := db.QueryRowContext(context.Background(), `SELECT version()`)
			if err := row.Scan(&s.BackendVersion); err != nil {
				l.Error("mysql.metadata.pool.openDB: failed to query MySQL version", zap.Error(err))
			}
		})
		if err != nil {
			l.Error("mysql.metadata.pool.openDB: failed to query MySQL version", zap.Error(err))
		}
	}

	return fsql.WrapDB(db, "", l), nil
}
