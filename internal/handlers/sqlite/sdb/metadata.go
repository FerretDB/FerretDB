package sdb

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (

	// Reserved prefix for database and collection names.
	reservedPrefix = "_ferretdb_"

	// Database metadata table name.
	dbMetadataTableName = reservedPrefix + "database_metadata"

	dbExtension = ".sqlite"
)

// NewMetadata returns instance of metadata
func NewMetadata(dbPath string) (*Metadata, error) {
	if dbPath == "" {
		return nil, errors.New("db path is empty")
	}

	return &Metadata{
		// TODO: might be passed as parameter.
		connPool: newConnPool(),
		dbPath:   dbPath,
		dbs:      make(map[string]*dbData),
	}, nil
}

type Metadata struct {
	dbPath string

	connPool *connPool

	mx  sync.Mutex
	dbs map[string]*dbData
}

type dbData struct {
	collections map[string]string
	// TODO: add indexes, etc.
}

func (m *Metadata) ListDatabases() ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	// short path
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

func (m *Metadata) ListCollections(ctx context.Context, database string) ([]string, error) {
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

func (m *Metadata) RemoveDatabase(database string) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	_, ok := m.dbs[database]
	if !ok {
		return errors.New("database not found")
	}

	delete(m.dbs, database)

	return nil
}

func (m *Metadata) RemoveCollection(database, collection string) error {
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

func (m *Metadata) load(ctx context.Context, dbName string) (*dbData, error) {
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

	var metadata = dbData{
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
