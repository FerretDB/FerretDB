package mdb

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/exp/maps"
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrDatabaseNotFound   = errors.New("database not found")
	ErrMalformedMetadata  = errors.New("malformed metadata")
)

const (
	dbExtension = ".sqlite"
)

func NewMetadata(dbPath string) (*Metadata, error) {
	if dbPath == "" {
		return nil, errors.New("db path is empty")
	}

	return &Metadata{
		dbPath: dbPath,
	}, nil
}

type Metadata struct {
	dbPath string

	mx  sync.Mutex
	dbs map[string]dbData
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

	m.dbs = make(map[string]dbData, len(dbs))

	for _, db := range dbs {
		conn, err := sql.Open("sqlite", db)
		if err != nil {
			return nil, err
		}

		_, err = conn.ExecContext(ctx, "")
		if err != nil {
			return nil, err
		}

		// load collections metadata here.
	}
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
