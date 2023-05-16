package mdb

import (
	"errors"
	"sync"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
)

var (
	ErrCollectionNotFound = errors.New("collection not found")
	ErrDatabaseNotFound   = errors.New("database not found")
	ErrMalformedMetadata  = errors.New("malformed metadata")
)

type MetadataStorage interface {
	CreateCollection(name string) error

	GetDatabasesList() ([]byte, error)

	RemoveCollection(database, collection string) error
	RemoveDatabase(database string) error
}

type Metadata struct {
	storage MetadataStorage

	mx  sync.Mutex
	dbs *types.Document
}

func (m *Metadata) GetDatabasesList() ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	// short path
	if m.dbs != nil {
		return m.dbs.Keys(), nil
	}

	list, err := m.storage.GetDatabasesList()
	if err != nil {
		return nil, err
	}

	dbs, err := sjson.Unmarshal(list)
	if err != nil {
		return nil, err
	}

	m.dbs = dbs

	return m.dbs.Keys(), nil
}

func (m *Metadata) GetCollectionsList(database string) ([]string, error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if !m.dbs.Has(database) {
		return nil, ErrDatabaseNotFound
	}

	dbMetadataDoc, err := m.dbs.Get(database)
	if err != nil {
		return nil, err
	}

	dbMetadata, ok := dbMetadataDoc.(*types.Document)
	if !ok {
		return nil, ErrMalformedMetadata
	}

	collMetadataDoc, err := dbMetadata.Get("collections")
	if err != nil {
		return nil, err
	}

	collMetadata, ok := collMetadataDoc.(*types.Document)
	if !ok {
		return nil, ErrMalformedMetadata
	}

	return collMetadata.Keys(), nil
}

func (m *Metadata) RemoveCollection(database, collection string) error {
	return nil
}

func (m *Metadata) RemoveDatabase(database string) error {
	return nil
}
