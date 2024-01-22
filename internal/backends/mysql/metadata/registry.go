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

package metadata

import (
	"context"
	"fmt"
	"hash/fnv"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/mysql/metadata/pool"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/sjson"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

const (
	// MySQL table name where FerretDB metadata is stored.
	metadataTableName = backends.ReservedPrefix + "database_metadata"

	// MySQL max table name length.
	maxTableNameLength = 64

	// MySQL max index name length.
	maxIndexNameLength = 64
)

// Parts of Prometheus metric names.
const (
	namespace = "ferretdb"
	subsystem = "mysql_metadata"
)

// specialCharacters are unsupported characters for MySQL table name that are replaced with `_`.
var specialCharacters = regexp.MustCompile("[^a-z][^a-z0-9_]*")

// Registry provides access to MySQL databases and collections information.
//
// Exported methods and [getPool] are safe for concurrent use. Other unexported methods are not.
//
// All methods should call [getPool] to check authentication.
// There is no authorization yet - if username/password is correct,
// all databases and collections are visible as far as Registry is concerned.
//
// Registry metadata is loaded upon first call by client, using [conninfo] in the context of the client.
//
//nolint:vet // for readability
type Registry struct {
	p *pool.Pool
	l *zap.Logger

	// rw protects colls but also acts like a global lock for the whole registry.
	// The latter effectively replaces transactions (see the mysql backend package description for more info).
	// One global lock should be replaced by more granular locks – one per database or even one per collection.
	// But that requires some redesign.
	// TODO https://github.com/FerretDB/FerretDB/issues/2755
	rw    sync.RWMutex
	colls map[string]map[string]*Collection // database name -> collection name -> collection
}

// NewRegistry creates a registry for the MySQL databases with a given base URI.
func NewRegistry(u string, l *zap.Logger, sp *state.Provider) (*Registry, error) {
	p, err := pool.New(u, l, sp)
	if err != nil {
		return nil, err
	}

	r := &Registry{
		p: p,
		l: l,
	}

	return r, nil
}

// Close closes the registry.
func (r *Registry) Close() {
	r.p.Close()
}

// getPool returns a pool of connections to MySQL database
// for the username/password combination in the context using [conninfo]
// (or any pool if authentication is bypassed)
//
// It loads metadata if it hasn't been loaded from the database yet.
//
// It acquires read lock to check metadata, if metadata is empty it acquires write lock
// to load metadata, so it is safe for concurrent use.
//
// All methods use this method to check authentication and load metadata.
func (r *Registry) getPool(ctx context.Context) (*fsql.DB, error) {
	connInfo := conninfo.Get(ctx)

	var p *fsql.DB

	if connInfo.BypassBackendAuth {
		if p = r.p.GetAny(); p == nil {
			return nil, lazyerrors.New("no connection pool")
		}
	} else {
		username, password := connInfo.Auth()

		var err error
		if p, err = r.p.Get(username, password); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	r.rw.RLock()
	if r.colls != nil {
		r.rw.RUnlock()
		return p, nil
	}
	r.rw.RUnlock()

	r.rw.Lock()
	defer r.rw.Unlock()

	dbNames, err := r.initDBs(ctx, p)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.colls = make(map[string]map[string]*Collection, len(dbNames))
	for _, dbName := range dbNames {
		if err = r.initCollections(ctx, dbName, p); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return p, nil
}

// initDBs returns a list of database names using schema information.
// It fetches existing schema (excluding ones reserved for MySQL),
// then finds and returns the schema that contains FerretDB metadata table.
func (r *Registry) initDBs(ctx context.Context, p *fsql.DB) ([]string, error) {
	q := strings.TrimSpace(`
		SELECT schema_name
		FROM information_schema.schemata
	`)

	rows, err := p.QueryContext(ctx, q)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var dbNames []string

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, lazyerrors.Error(err)
		}

		// schema created by MySQL can be used as a FerretDB database,
		// but if it does not contain FerretDB metadata table, it is not used by FerretDB
		q := strings.TrimSpace(`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = ? AND table_name = ?
			)
		`)

		var exist bool
		if err = p.QueryRowContext(ctx, q, dbName, metadataTableName).Scan(&exist); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if exist {
			dbNames = append(dbNames, dbName)
		}
	}

	return dbNames, nil
}

// initCollection loads collection metadata from the database during initialization.
func (r *Registry) initCollections(ctx context.Context, dbName string, p *fsql.DB) error {
	defer observability.FuncCall(ctx)()

	q := fmt.Sprintf(
		`SELECT %s FROM %s.%s`,
		DefaultColumn,
		dbName, metadataTableName,
	)

	rows, err := p.QueryContext(ctx, q)
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer rows.Close()

	colls := map[string]*Collection{}

	for rows.Next() {
		var c Collection

		if err = rows.Scan(&c); err != nil {
			return lazyerrors.Error(err)
		}
		colls[c.Name] = &c
	}

	if err = rows.Err(); err != nil {
		return lazyerrors.Error(err)
	}

	r.colls[dbName] = colls

	return nil
}

// DatabaseList returns a sorted list of existing databases.
//
// If the user is not authenticated, it returns an error.
func (r *Registry) DatabaseList(ctx context.Context) ([]string, error) {
	defer observability.FuncCall(ctx)()

	_, err := r.getPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.RLock()
	defer r.rw.RUnlock()

	res := maps.Keys(r.colls)
	sort.Strings(res)

	return res, nil
}

// DatabaseGetExisting returns a connection to existing database or nil if it doesn't exist.
//
// If the user is not authenticated, it returns error.
func (r *Registry) DatabaseGetExisting(ctx context.Context, dbName string) (*fsql.DB, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.RLock()
	defer r.rw.RUnlock()

	db := r.colls[dbName]
	if db == nil {
		return nil, nil
	}

	return p, nil
}

// DatabaseGetOrCreate returns a connection to existing database or newly created database.
//
// The dbName must be a validated database name.
//
// If the user is not authenticated, it returns error.
func (r *Registry) DatabaseGetOrCreate(ctx context.Context, dbName string) (*fsql.DB, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.databaseGetOrCreate(ctx, p, dbName)
}

// databaseGetOrCreate returns a connection to the existing database or newly created database.
//
// The dbName must be a validated database name.
//
// It does not hold the lock.
func (r *Registry) databaseGetOrCreate(ctx context.Context, p *fsql.DB, dbName string) (*fsql.DB, error) {
	defer observability.FuncCall(ctx)()

	db := r.colls[dbName]
	if db != nil {
		return p, nil
	}

	q := fmt.Sprintf(
		`CREATE SCHEMA IF NOT EXISTS %s`,
		dbName,
	)

	var err error
	if _, err = p.ExecContext(ctx, q); err != nil {
		return nil, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`CREATE TABLE %s.%s (%s json)`,
		dbName, metadataTableName,
		DefaultColumn,
	)

	if _, err = p.ExecContext(ctx, q); err != nil {
		_, _ = r.databaseDrop(ctx, p, dbName)
		return nil, lazyerrors.Error(err)
	}

	// json columns cannot be indexed directly in MySQL. A workaround for this is done by creating a
	// generated column that extracts information that should be indexed.
	//
	// https://dev.mysql.com/doc/refman/5.7/en/create-table-secondary-indexes.html#json-column-indirect-index
	q = fmt.Sprintf(
		`ALTER TABLE %s.%s
		 ADD COLUMN %s VARCHAR(255) GENERATED ALWAYS AS ((%s)) STORED,
    	 ADD COLUMN %s VARCHAR(255) GENERATED ALWAYS AS ((%s)) STORED
		`,
		dbName, metadataTableName,
		IDIndexColumn, IDColumn,
		TableIndexColumn, DefaultColumn+"->'$.table'",
	)

	if _, err = p.ExecContext(ctx, q); err != nil {
		_, _ = r.databaseDrop(ctx, p, dbName)
		return nil, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`CREATE UNIQUE INDEX %s ON %s.%s (%s)`,
		metadataTableName+"_id_idx",
		dbName, metadataTableName,
		IDIndexColumn,
	)

	if _, err = p.ExecContext(ctx, q); err != nil {
		_, _ = r.databaseDrop(ctx, p, dbName)
		return nil, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`CREATE UNIQUE INDEX %s ON %s.%s (%s)`,
		metadataTableName+"_table_idx",
		dbName, metadataTableName,
		TableIndexColumn,
	)

	if _, err = p.ExecContext(ctx, q); err != nil {
		_, _ = r.databaseDrop(ctx, p, dbName)
		return nil, lazyerrors.Error(err)
	}

	r.colls[dbName] = map[string]*Collection{}

	return p, nil
}

// DatabaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
// If database does not exist, (false, nil) is returned.
//
// If user is not authenticated, it returns error.
func (r *Registry) DatabaseDrop(ctx context.Context, dbName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.databaseDrop(ctx, p, dbName)
}

// DatabaseDrop drops the database.
//
// Returned boolean value indicates whether the database was dropped.
// If database does not exist, (false, nil) is returned.
//
// It does not hold the lock.
func (r *Registry) databaseDrop(ctx context.Context, p *fsql.DB, dbName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	db := r.colls[dbName]
	if db == nil {
		return false, nil
	}

	// fix cascade delete for mysql
	q := fmt.Sprintf(
		`DROP DATABASE %s`,
		dbName,
	)

	if _, err := p.ExecContext(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	delete(r.colls, dbName)

	return true, nil
}

// CollectionList returns a sorted copy of collections in the database.
//
// If database does not exist, no error is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) CollectionList(ctx context.Context, dbName string) ([]*Collection, error) {
	defer observability.FuncCall(ctx)()

	if _, err := r.getPool(ctx); err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.RLock()
	defer r.rw.RUnlock()

	db := r.colls[dbName]
	if db == nil {
		return nil, nil
	}

	res := make([]*Collection, 0, len(r.colls[dbName]))
	for _, c := range r.colls[dbName] {
		res = append(res, c.deepCopy())
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })

	return res, nil
}

// CollectionCreateParams contains parameters for CollectionCreate.
type CollectionCreateParams struct {
	DBName          string
	Name            string
	CappedSize      int64
	CappedDocuments int64
}

// Capped returns true if capped collection creation is requested.
func (cpp *CollectionCreateParams) Capped() bool {
	return cpp.CappedSize > 0 // TODO https://github.com/FerretDB/FerretDB/issues/3631
}

// CollectionCreate creates a collection in the database.
// Database will be created automatically if needed.
//
// Returned boolean value indicates whether the collection was created.
// If collection already exists, (false, nil) is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) CollectionCreate(ctx context.Context, params *CollectionCreateParams) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.collectionCreate(ctx, p, params)
}

// collectionCreate creates a collection in the database.
// Database will be created automatically if needed.
//
// Returned boolean value indicates whether the collection was created.
// If collection already exists, (false, nil) is returned.
//
// It does not hold the lock.
func (r *Registry) collectionCreate(ctx context.Context, p *fsql.DB, params *CollectionCreateParams) (bool, error) {
	defer observability.FuncCall(ctx)()

	dbName, collectionName := params.DBName, params.Name

	_, err := r.databaseGetOrCreate(ctx, p, dbName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	colls := r.colls[dbName]
	if colls != nil && colls[collectionName] != nil {
		return false, nil
	}

	h := fnv.New32a()
	must.NotFail(h.Write([]byte(collectionName)))
	s := h.Sum32()

	var tableName string
	list := maps.Values(colls)

	for {
		tableName = specialCharacters.ReplaceAllString(strings.ToLower(collectionName), "_")

		suffixHash := fmt.Sprintf("_%08x", s)
		if l := maxTableNameLength - len(suffixHash); len(tableName) > l {
			tableName = tableName[:l]
		}

		tableName = fmt.Sprintf("%s%s", tableName, suffixHash)

		if !slices.ContainsFunc(list, func(c *Collection) bool { return c.TableName == tableName }) {
			break
		}

		// table already exists, generate a new table by incrementing the hash
		s++
	}

	c := &Collection{
		Name:            collectionName,
		UUID:            uuid.NewString(),
		TableName:       tableName,
		CappedSize:      params.CappedSize,
		CappedDocuments: params.CappedDocuments,
	}

	q := fmt.Sprintf(`CREATE TABLE %s.%s (`, dbName, tableName)

	if params.Capped() {
		q += fmt.Sprintf(`%s bigint PRIMARY KEY, `, RecordIDColumn)
	}

	q += fmt.Sprintf(`%s json)`, DefaultColumn)

	if _, err = p.ExecContext(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`INSERT INTO %s.%s (%s) VALUES (?)`,
		dbName, metadataTableName,
		DefaultColumn,
	)

	if _, err = p.ExecContext(ctx, q, c); err != nil {
		q = fmt.Sprintf(`DROP TABLE %s.%s`, dbName, tableName)
		_, _ = p.ExecContext(ctx, q)

		return false, lazyerrors.Error(err)
	}

	if r.colls[dbName] == nil {
		r.colls[dbName] = map[string]*Collection{}
	}
	r.colls[dbName][collectionName] = c

	err = r.indexesCreate(ctx, p, dbName, collectionName, []IndexInfo{{
		Name:   "_id_",
		Key:    []IndexKeyPair{{Field: "_id"}},
		Unique: true,
	}})
	if err != nil {
		_, _ = r.collectionDrop(ctx, p, dbName, collectionName)
		return false, lazyerrors.Error(err)
	}

	return true, nil
}

// CollectionGet returns a copy of collection metadata.
// It can be safely modified by a caller.
//
// If database or collection does not exist, nil is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) CollectionGet(ctx context.Context, dbName, collectionName string) (*Collection, error) {
	defer observability.FuncCall(ctx)()

	if _, err := r.getPool(ctx); err != nil {
		return nil, lazyerrors.Error(err)
	}

	r.rw.RLock()
	defer r.rw.RUnlock()

	return r.collectionGet(dbName, collectionName), nil
}

// collectionGet returns a copy of collection metadata.
// It can be safely modified by a caller.
//
// If database or collection does not exist, nil is returned.
//
// It does not hold the lock.
func (r *Registry) collectionGet(dbName, collectionName string) *Collection {
	colls := r.colls[dbName]
	if colls == nil {
		return nil
	}

	return colls[collectionName].deepCopy()
}

// CollectionDrop drops a collection in the database.
//
// Returned boolean value indicates whether the collection was dropped.
// If database or collection did not exist, (false, nil) is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) CollectionDrop(ctx context.Context, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.collectionDrop(ctx, p, dbName, collectionName)
}

// collectionDrop drops a collection in the database.
//
// Returned boolean value indicates whether the collection was dropped.
// If database or collection did not exist, (false, nil) is returned.
//
// It does not hold the lock.
func (r *Registry) collectionDrop(ctx context.Context, p *fsql.DB, dbName, collectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	db := r.colls[dbName]
	if db == nil {
		return false, nil
	}

	c := r.collectionGet(dbName, collectionName)
	if c == nil {
		return false, nil
	}

	arg, err := sjson.MarshalSingleValue(c.Name)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	q := fmt.Sprintf(
		`DROP TABLE %s.%s`,
		dbName, c.TableName,
	)

	if _, err := p.ExecContext(ctx, q); err != nil {
		return false, lazyerrors.Error(err)
	}

	q = fmt.Sprintf(
		`DELETE FROM %s.%s WHERE %s IN (?)`,
		dbName, metadataTableName,
		IDIndexColumn,
	)

	if _, err := p.ExecContext(ctx, q, string(arg)); err != nil {
		return false, lazyerrors.Error(err)
	}

	delete(r.colls[dbName], collectionName)

	return true, nil
}

// CollectionRename renames a collection in the database.
//
// The collection name is updated, but the original table name is kept.
//
// Returned boolean value indicates whether the collection was renamed.
// If database or collection did not exist, (false, nil) is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) CollectionRename(ctx context.Context, dbName, oldCollectionName, newCollectionName string) (bool, error) {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	db := r.colls[dbName]
	if db == nil {
		return false, nil
	}

	c := r.collectionGet(dbName, oldCollectionName)
	if c == nil {
		return false, nil
	}

	c.Name = newCollectionName

	b, err := sjson.Marshal(c.marshal())
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	arg, err := sjson.MarshalSingleValue(oldCollectionName)
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	q := fmt.Sprintf(
		`UPDATE %s.%s SET %s = ? WHERE %s = ?`,
		dbName, metadataTableName,
		DefaultColumn,
		IDIndexColumn,
	)

	if _, err := p.ExecContext(ctx, q, string(b), arg); err != nil {
		return false, lazyerrors.Error(err)
	}

	r.colls[dbName][newCollectionName] = c
	delete(r.colls[dbName], oldCollectionName)

	return true, nil
}

// IndexesCreate creates indexes in the collection.
//
// Existing indexes with given names are ignored.
//
// If the user is not authenticated, it returns error.
func (r *Registry) IndexesCreate(ctx context.Context, dbName, collectionName string, indexes []IndexInfo) error {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.indexesCreate(ctx, p, dbName, collectionName, indexes)
}

// indexesCreate creates indexes in the collection.
// Existing indexes with given names are ignored.
//
// It does not hold the lock.
func (r *Registry) indexesCreate(ctx context.Context, p *fsql.DB, dbName, collectionName string, indexes []IndexInfo) error {
	defer observability.FuncCall(ctx)()

	_, err := r.collectionCreate(ctx, p, &CollectionCreateParams{DBName: dbName, Name: collectionName})
	if err != nil {
		return lazyerrors.Error(err)
	}

	db := r.colls[dbName]
	if db == nil {
		panic("database does not exist")
	}

	c := r.collectionGet(dbName, collectionName)
	if c == nil {
		panic("collection does not exist")
	}

	allIndexes := make(map[string]string, len(db))      // to check if the index already exists
	allMySQLIndexes := make(map[string]string, len(db)) // to ensure there are no indexes with the same name in the MySQL schema

	for _, coll := range db {
		for _, index := range coll.Indexes {
			allIndexes[index.Name] = coll.Name
			allMySQLIndexes[index.Index] = coll.Name
		}
	}

	created := make([]string, 0, len(indexes))

	for _, index := range indexes {
		if coll, ok := allIndexes[index.Name]; ok && coll == collectionName {
			continue
		}

		tableNamePart := c.TableName
		tableNamePartMax := maxIndexNameLength/2 - 1 // 1 for the separator between table name and index name

		if len(tableNamePart) > tableNamePartMax {
			tableNamePart = tableNamePart[:tableNamePartMax]
		}

		indexNamePart := specialCharacters.ReplaceAllString(strings.ToLower(index.Name), "_")

		h := fnv.New32a()
		must.NotFail(h.Write([]byte(index.Name)))
		s := h.Sum32()

		var mysqlIndexName string

		for {
			suffixHash := fmt.Sprintf("_%08x_idx", s)
			if l := maxIndexNameLength/2 - len(suffixHash); len(indexNamePart) > l {
				indexNamePart = indexNamePart[:l]
			}

			mysqlIndexName = fmt.Sprintf("%s_%s%s", tableNamePart, indexNamePart, suffixHash)

			// indexes must be unique across the whole database, so we check for duplicates for all other collections.
			_, duplicate := allMySQLIndexes[mysqlIndexName]

			if !duplicate {
				break
			}

			s++
		}

		index.Index = mysqlIndexName

		q := `
			SELECT column_name FROM INFORMATION_SCHEMA.COLUMNS WHERE table_schema = ? AND table_name = ?
		`

		var allColumns []string

		var rows *fsql.Rows

		rows, err = p.QueryContext(ctx, q, dbName, c.TableName)
		if err != nil {
			return lazyerrors.Error(err)
		}
		defer rows.Close()

		for rows.Next() {
			var c string

			if err = rows.Scan(&c); err != nil {
				return lazyerrors.Error(err)
			}
			allColumns = append(allColumns, c)
		}

		if err = rows.Err(); err != nil {
			return lazyerrors.Error(err)
		}

		q = "ALTER TABLE %s.%s"

		columns := make([]string, len(index.Key))

		for i, key := range index.Key {
			columnName := strings.ReplaceAll(key.Field, ".", "_")

			// ensure that the column hasn't already been extracted
			if !slices.Contains(allColumns, columnName) {
				q += fmt.Sprintf(
					` ADD COLUMN %s VARCHAR(255) GENERATED ALWAYS AS ((%s->'%s')) STORED`,
					columnName,
					DefaultColumn,
					"$."+key.Field,
				)

				if i != len(index.Key)-1 {
					q += ","
				}
			}

			columns[i] = key.Field

			if key.Descending {
				columns[i] += " DESC"
			}
		}

		q = fmt.Sprintf(
			q,
			dbName, c.TableName,
		)

		if _, err = p.ExecContext(ctx, q); err != nil {
			return lazyerrors.Error(err)
		}

		q = "CREATE "

		if index.Unique {
			q += "UNIQUE "
		}

		q += "INDEX %s ON %s.%s (%s)"

		q = fmt.Sprintf(
			q,
			index.Index,
			dbName, c.TableName,
			strings.Join(columns, ", "),
		)

		if _, err = p.ExecContext(ctx, q); err != nil {
			_ = r.indexesDrop(ctx, p, dbName, collectionName, created)
			return lazyerrors.Error(err)
		}

		created = append(created, index.Name)
		c.Indexes = append(c.Indexes, index)
		allIndexes[index.Name] = collectionName
		allMySQLIndexes[index.Index] = collectionName
	}

	b, err := sjson.Marshal(c.marshal())
	if err != nil {
		return lazyerrors.Error(err)
	}

	arg, err := sjson.MarshalSingleValue(collectionName)
	if err != nil {
		return lazyerrors.Error(err)
	}

	q := fmt.Sprintf(
		`UPDATE %s.%s SET %s = ? WHERE %s = ?`,
		dbName, metadataTableName,
		DefaultColumn,
		IDIndexColumn,
	)

	if _, err := p.ExecContext(ctx, q, string(b), string(arg)); err != nil {
		return lazyerrors.Error(err)
	}

	r.colls[dbName][collectionName] = c

	return nil
}

// IndexesDrop removes given connection's indexes.
//
// Non-existing indexes are ignored.
//
// If database or collection does not exist, nil is returned.
//
// If the user is not authenticated, it returns error.
func (r *Registry) IndexesDrop(ctx context.Context, dbName, collectionName string, indexNames []string) error {
	defer observability.FuncCall(ctx)()

	p, err := r.getPool(ctx)
	if err != nil {
		return lazyerrors.Error(err)
	}

	r.rw.Lock()
	defer r.rw.Unlock()

	return r.indexesDrop(ctx, p, dbName, collectionName, indexNames)
}

// indexesDrop removes given connection's indexes.
//
// Non-existing indexes are ignored.
//
// If database or collection does not exist, nil is returned.
//
// It does not hold the lock.
func (r *Registry) indexesDrop(ctx context.Context, p *fsql.DB, dbName, collectionName string, indexNames []string) error {
	defer observability.FuncCall(ctx)()

	// check if the collection exists
	c := r.collectionGet(dbName, collectionName)
	if c == nil {
		return nil
	}

	// find the collection in the list of collections
	for _, name := range indexNames {
		i := slices.IndexFunc(c.Indexes, func(i IndexInfo) bool { return name == i.Name })
		if i < 0 {
			continue
		}

		q := fmt.Sprintf("DROP INDEX %s.%s", dbName, c.Indexes[i].Index)
		if _, err := p.ExecContext(ctx, q); err != nil {
			return lazyerrors.Error(err)
		}

		c.Indexes = slices.Delete(c.Indexes, i, i+1)
	}

	b, err := sjson.Marshal(c.marshal())
	if err != nil {
		return lazyerrors.Error(err)
	}

	arg, err := sjson.MarshalSingleValue(collectionName)
	if err != nil {
		return lazyerrors.Error(nil)
	}

	q := fmt.Sprintf(
		`UPDATE %s.%s SET %s = ? WHERE %s = ?`,
		dbName, metadataTableName,
		DefaultColumn,
		IDColumn,
	)

	if _, err := p.ExecContext(ctx, q, string(b), arg); err != nil {
		return lazyerrors.Error(err)
	}

	r.colls[dbName][collectionName] = c

	return nil
}

// Describe implements prometheus.Collector.
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(r, ch)
}

// Collect implements prometheus.Collector.
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	r.p.Collect(ch)

	r.rw.RLock()
	defer r.rw.RUnlock()

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "databases"),
			"The current number of database in the registry.",
			nil, nil,
		),
		prometheus.GaugeValue,
		float64(len(r.colls)),
	)

	for db, colls := range r.colls {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, "collections"),
				"The current number of collections in the registry.",
				[]string{"db"}, nil,
			),
			prometheus.GaugeValue,
			float64(len(colls)),
			db,
		)
	}
}

// check interfaces
var (
	_ prometheus.Collector = (*Registry)(nil)
)
