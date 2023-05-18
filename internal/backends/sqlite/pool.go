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

package sqlite

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sync"

	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// newConnPool creates a new connection pool.
func newConnPool() *connPool {
	pool := &connPool{
		mx:    sync.Mutex{},
		dbs:   map[string]*sql.DB{},
		token: resource.NewToken(),
	}

	resource.Track(pool, pool.token)

	return pool
}

// connPool is a pool of database connections.
type connPool struct { //nolint:vet // for readability
	mx  sync.Mutex
	dbs map[string]*sql.DB

	token *resource.Token
	dir   string
}

// DB returns a database connection for the given name.
func (c *connPool) DB(name string) (*sql.DB, error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	path := filepath.Join(c.dir, name+dbExtension)

	if db, ok := c.dbs[path]; ok {
		return db, nil
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	c.dbs[path] = db

	return db, nil
}

// Close closes all database connections.
func (c *connPool) Close() error {
	var errs error

	c.mx.Lock()
	defer c.mx.Unlock()

	for _, conn := range c.dbs {
		if err := conn.Close(); err != nil {
			errors.Join(err)
		}
	}

	resource.Untrack(c, c.token)

	return errs
}
