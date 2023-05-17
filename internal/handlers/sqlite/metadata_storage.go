package sqlite

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (

	// Reserved prefix for database and collection names.
	reservedPrefix = "_ferretdb_"

	// Database metadata table name.
	dbMetadataTableName = reservedPrefix + "database_metadata"

	dbExtension = ".sqlite"
)

// newMetadataStorage returns instance of metadata storage.
func newMetadataStorage(dbPath string) (*metadataStorage, error) {
	if dbPath == "" {
		return nil, errors.New("db path is empty")
	}

	return &metadataStorage{
		// TODO: might be passed as parameter.
		connPool: newConnPool(),
		dbPath:   dbPath,
		dbs:      make(map[string]*dbInfo),
	}, nil
}

// metadataStorage provide access to database metadata.
type metadataStorage struct {
	dbPath string

	connPool *connPool

	mx  sync.Mutex
	dbs map[string]*dbInfo
}

type dbInfo struct {
	collections map[string]string
	// TODO: add indexes, etc.
}

// ListDatabases list database names.
func (m *metadataStorage) ListDatabases() ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.dbs != nil {
		return maps.Keys(m.dbs), nil
	}

	var dbs []string

	err := filepath.WalkDir(m.dbPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == dbExtension {
			dbs = append(dbs, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// fill in all databases that were found.
	for _, db := range dbs {
		m.dbs[db] = nil
	}

	return dbs, nil
}

// ListCollections list collection names for given database.
func (m *metadataStorage) ListCollections(ctx context.Context, database string) ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return nil, errors.New("database not found")
	}

	// no metadata about collections loaded, load now
	if db == nil {
		var err error

		db, err = m.load(ctx, database)
		if err != nil {
			return nil, err
		}

		m.dbs[database] = db
	}

	return maps.Keys(db.collections), nil
}

// RemoveDatabase removes database metadata.
// It does not remove database file.
func (m *metadataStorage) RemoveDatabase(database string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	_, ok := m.dbs[database]
	if !ok {
		return errors.New("database not found")
	}

	delete(m.dbs, database)

	return nil
}

// RemoveCollection removes collection metadata.
// It does not physically remove collection from database.
func (m *metadataStorage) RemoveCollection(database, collection string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return errors.New("database not found")
	}

	_, ok = db.collections[collection]
	if !ok {
		return errors.New("collection not found")
	}

	delete(db.collections, collection)

	return nil
}

func (m *metadataStorage) CreateDatabase(ctx context.Context, database string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	_, ok := m.dbs[database]
	if ok {
		return errors.New("database already exists")
	}

	m.dbs[database] = nil

	return nil
}

// CreateCollection saves collection metadata to database file
func (m *metadataStorage) CreateCollection(ctx context.Context, database, collection string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	db, ok := m.dbs[database]
	if !ok {
		return errors.New("database not found")
	}

	_, ok = db.collections[collection]
	if ok {
		return errors.New("collection already exists")
	}

	tableName := mangleCollection(collection)

	err := m.saveCollection(ctx, database, collection, tableName)
	if err != nil {
		return err
	}

	db.collections[collection] = tableName

	return nil
}

// load loads database metadata from database file.
func (m *metadataStorage) load(ctx context.Context, dbName string) (*dbInfo, error) {
	conn, err := m.connPool.DB(dbName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT sjson FROM %s", dbMetadataTableName)

	result, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	// TODO: check error
	defer result.Close()

	var metadata = dbInfo{
		collections: make(map[string]string),
	}

	for result.Next() {
		var rawBytes []byte

		err = result.Scan(rawBytes)
		if err != nil {
			return nil, err
		}

		doc, err := sjson.Unmarshal(rawBytes)
		if err != nil {
			return nil, errors.New("failed to unmarshal collection metadata")
		}

		// TODO: proper errors check
		collName := must.NotFail(doc.Get("collection")).(string)
		tableName := must.NotFail(doc.Get("table")).(string)

		metadata.collections[collName] = tableName
	}

	return &metadata, nil
}

// saveCollection saves collection metadata to database file.
func (m *metadataStorage) saveCollection(ctx context.Context, dbName, collName, tableName string) error {
	conn, err := m.connPool.DB(dbName)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO %s (sjson) VALUES (?)", dbMetadataTableName)

	bytes, err := sjson.Marshal(must.NotFail(types.NewDocument("collection", collName, "table", tableName)))
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, query, bytes)
	if err != nil {
		return err
	}

	return nil
}

// mangleCollection mangles collection name to table name.
// TODO: implement proper mangle if needed.
func mangleCollection(name string) string {
	return name
}
