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

// Package pool provides access to SQLite database connections.
//
// It should be used only by the metadata package.
package pool

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// filenameExtension represents SQLite database filename extension.
const filenameExtension = ".sqlite"

// Pool provides access to SQLite database connections.
//
//nolint:vet // for readability
type Pool struct {
	uri *url.URL
	l   *zap.Logger

	rw  sync.RWMutex
	dbs map[string]*db

	token *resource.Token
}

// New creates a pool for SQLite databases in the given directory.
func New(u string, l *zap.Logger) (*Pool, error) {
	uri, err := validateURI(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQLite URI %q: %s", u, err)
	}

	matches, err := filepath.Glob(filepath.Join(uri.Path, "*"+filenameExtension))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	p := &Pool{
		uri:   uri,
		l:     l,
		dbs:   make(map[string]*db, len(matches)),
		token: resource.NewToken(),
	}

	resource.Track(p, p.token)

	for _, f := range matches {
		name := p.databaseName(f)
		uri := p.databaseURI(name)

		p.l.Debug("Opening existing database.", zap.String("name", name), zap.String("uri", uri))

		db, err := openDB(uri)
		if err != nil {
			p.Close()
			return nil, lazyerrors.Error(err)
		}

		p.dbs[name] = db
	}

	return p, nil
}

// validateURI checks given URI value and returns parsed URL.
// URI should contain 'file' scheme and point to an existing directory.
// Path should end with '/'. Authority should be empty or absent.
//
// Returned URL contains path in both Path and Opaque to make String() method work correctly.
func validateURI(u string) (*url.URL, error) {
	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	if uri.Scheme != "file" {
		return nil, fmt.Errorf(`expected "file:" schema, got %q`, uri.Scheme)
	}

	if uri.User != nil {
		return nil, fmt.Errorf(`expected empty user info, got %q`, uri.User)
	}

	if uri.Host != "" {
		return nil, fmt.Errorf(`expected empty host, got %q`, uri.Host)
	}

	if uri.Path == "" && uri.Opaque != "" {
		uri.Path = uri.Opaque
	}
	uri.Opaque = uri.Path

	if !strings.HasSuffix(uri.Path, "/") {
		return nil, fmt.Errorf(`expected path ending with "/", got %q`, uri.Path)
	}

	fi, err := os.Stat(uri.Path)
	if err != nil {
		return nil, fmt.Errorf(`%q should be an existing directory, got %s`, uri.Path, err)
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf(`%q should be an existing directory`, uri.Path)
	}

	return uri, nil
}

// databaseName returns database name for given database file path.
func (p *Pool) databaseName(databaseFile string) string {
	return strings.TrimSuffix(filepath.Base(databaseFile), filenameExtension)
}

// databaseURI returns SQLite URI for the given database name.
func (p *Pool) databaseURI(databaseName string) string {
	dbURI := *p.uri
	dbURI.Path = path.Join(dbURI.Path, databaseName+filenameExtension)
	dbURI.Opaque = dbURI.Path

	return dbURI.String()
}

// databaseFile returns database file path for the given database name.
func (p *Pool) databaseFile(databaseName string) string {
	return filepath.Join(p.uri.Path, databaseName+filenameExtension)
}

// Close closes all databases in the pool and frees all resources.
func (p *Pool) Close() {
	p.rw.Lock()
	defer p.rw.Unlock()

	for _, db := range p.dbs {
		_ = db.Close()
	}

	p.dbs = nil

	resource.Untrack(p, p.token)
}

// List returns a sorted list of database names in the pool.
func (p *Pool) List(ctx context.Context) []string {
	p.rw.RLock()
	defer p.rw.RUnlock()

	res := maps.Keys(p.dbs)
	slices.Sort(res)

	return res
}

// GetExisting returns an existing database connection by name, or nil.
func (p *Pool) GetExisting(ctx context.Context, name string) *sql.DB {
	p.rw.RLock()
	defer p.rw.RUnlock()

	db := p.dbs[name]
	if db == nil {
		return nil
	}

	return db.sqlDB
}

// GetOrCreate returns an existing database connection by name, or creates a new one.
//
// Returned boolean value indicates whether the connection was created.
func (p *Pool) GetOrCreate(ctx context.Context, name string) (*sql.DB, bool, error) {
	sqlDB := p.GetExisting(ctx, name)
	if sqlDB != nil {
		return sqlDB, false, nil
	}

	p.rw.Lock()
	defer p.rw.Unlock()

	// it might have been created by a concurrent call
	if db := p.dbs[name]; db != nil {
		return db.sqlDB, false, nil
	}

	uri := p.databaseURI(name)
	db, err := openDB(uri)
	if err != nil {
		return nil, false, lazyerrors.Errorf("%s: %w", uri, err)
	}

	p.l.Debug("Database created.", zap.String("name", name), zap.String("uri", uri))

	p.dbs[name] = db

	return db.sqlDB, true, nil
}

// Drop closes and removes a database by name.
//
// It does nothing if the database does not exist.
//
// Returned boolean value indicates whether the database was removed.
func (p *Pool) Drop(ctx context.Context, name string) bool {
	p.rw.Lock()
	defer p.rw.Unlock()

	db, ok := p.dbs[name]
	if !ok {
		return false
	}

	_ = db.Close()
	_ = os.Remove(p.databaseFile(name))
	delete(p.dbs, name)

	p.l.Debug("Database dropped.", zap.String("name", name))

	return true
}

// InTransaction wraps the given function f in a SQLite transaction.
//
// If f returns an error, the transaction is rolled back.
//
// Passed context will be used for transaction.
// Context cancellation DOES rollback the transaction.
func InTransaction(ctx context.Context, db *sql.DB, f func(*sql.Tx) error) (err error) {
	var tx *sql.Tx

	if tx, err = db.BeginTx(ctx, nil); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	if err = f(tx); err != nil {
		err = lazyerrors.Error(err)
		_ = tx.Rollback()

		return
	}

	if err = tx.Commit(); err != nil {
		err = lazyerrors.Error(err)
		_ = tx.Rollback()

		return
	}

	return
}
