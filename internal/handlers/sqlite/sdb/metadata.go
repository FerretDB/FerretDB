package sdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrDatabaseNotFound   = errors.New("database not found")
	ErrMalformedMetadata  = errors.New("malformed metadata")
)

const (

	// Reserved prefix for database and collection names.
	reservedPrefix = "_ferretdb_"

	// Database metadata table name.
	dbMetadataTableName = reservedPrefix + "database_metadata"

	dbExtension = ".sqlite"
)

func NewMetadata(dbPath string) (*Metadata, error) {
	if dbPath == "" {
		return nil, errors.New("db path is empty")
	}

	return &Metadata{
		dbPath: dbPath,
		dbs:    make(map[string]*dbData),
	}, nil
}

type Metadata struct {
	dbPath string

	mx  sync.Mutex
	dbs map[string]*dbData
}

type dbData struct {
	collections map[string]string
	// indexes, etc.
}

func (m *Metadata) GetDatabasesList(ctx context.Context) ([]string, error) {
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

	for _, dbName := range dbs {
		conn, err := sql.Open("sqlite", dbName)
		if err != nil {
			return nil, err
		}

		query := fmt.Sprintf("SELECT sjson FROM %s", dbMetadataTableName)

		result, err := conn.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}

		var doc string

		err = result.Scan(&doc)
		if err != nil {
			return nil, err
		}

		metadata, err := m.documentToMetadata(doc)
		if err != nil {
			return nil, err
		}

		m.dbs[dbName] = metadata
	}

	return nil, nil
}

func (m *Metadata) GetCollectionsList(database string) ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	return nil, nil
}

func (m *Metadata) RemoveCollection(database, collection string) error {
	return nil
}

func (m *Metadata) RemoveDatabase(database string) error {
	return nil
}

func (m *Metadata) documentToMetadata(raw []byte) (*dbData, error) {
	doc, err := sjson.Unmarshal(raw)
	if err != nil {
		return nil, err
	}

	if !doc.Has("_id") {
		return nil, ErrMalformedMetadata
	}

	doc.Get("_id")
}
