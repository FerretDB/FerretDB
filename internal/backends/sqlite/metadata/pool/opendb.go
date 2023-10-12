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

	"go.uber.org/zap"
	_ "modernc.org/sqlite" // register database/sql driver

	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

// openDB opens existing database or creates a new one.
//
// All valid FerretDB database names are valid SQLite database names / file names,
// so no validation is needed.
// One exception is very long full path names for the filesystem,
// but we don't check it.
func openDB(name, uri string, memory bool, l *zap.Logger, sp *state.Provider) (*fsql.DB, error) {
	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)

	// Each connection to in-memory database uses its own database.
	// See https://www.sqlite.org/inmemorydb.html.
	// We don't want that.
	if memory {
		db.SetMaxIdleConns(1)
		db.SetMaxOpenConns(1)
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, lazyerrors.Error(err)
	}

	if sp.Get().HandlerVersion == "" {
		err := sp.Update(func(s *state.State) {
			row := db.QueryRowContext(context.Background(), "SELECT sqlite_version()")
			if err := row.Scan(&s.HandlerVersion); err != nil {
				l.Error("sqlite.metadata.pool.openDB: failed to query SQLite version", zap.Error(err))
			}
		})
		if err != nil {
			l.Error("sqlite.metadata.pool.openDB: failed to update state", zap.Error(err))
		}
	}

	return fsql.WrapDB(db, name, l), nil
}
