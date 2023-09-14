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

// Package metadata provides access to databases and collections information.
package metadata

import (
	"errors"
	"fmt"
	"slices"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// DefaultColumn is a column name for all fields.
	DefaultColumn = "_jsonb"
)

// Collection represents collection metadata.
type Collection struct {
	Name      string `json:"_id"`
	TableName string `json:"table"`
	Settings  Settings
}

// deepCopy returns a deep copy.
func (c *Collection) deepCopy() *Collection {
	if c == nil {
		return nil
	}

	return &Collection{
		Name:      c.Name,
		TableName: c.TableName,
		Settings:  c.Settings.deepCopy(),
	}
}

// Settings represents collection settings.
type Settings struct {
	Indexes []IndexInfo `json:"indexes"`
}

// IndexInfo represents information about a single index.
type IndexInfo struct {
	PgIndex string
	Name    string         `json:"name"`
	Key     []IndexKeyPair `json:"key"`
	Unique  bool           `json:"unique"`
}

// IndexKeyPair consists of a field name and a sort order that are part of the index.
type IndexKeyPair struct {
	Field      string `json:"field"`
	Descending bool   `json:"descending"`
}

// deepCopy returns a deep copy.
func (s Settings) deepCopy() Settings {
	indexes := make([]IndexInfo, len(s.Indexes))

	for i, index := range s.Indexes {
		indexes[i] = IndexInfo{
			Name:   index.Name,
			Key:    slices.Clone(index.Key),
			Unique: index.Unique,
		}
	}

	return Settings{
		Indexes: indexes,
	}
}

// Marshal returns [*types.Document] for that collection.
func (c *Collection) Marshal() *types.Document {
	indexes := types.MakeArray(len(c.Settings.Indexes))

	for _, idx := range c.Settings.Indexes {
		keyDoc := types.MakeDocument(len(idx.Key))
		for _, pair := range idx.Key {
			order := int32(1)
			if pair.Descending {
				order = int32(-1)
			}
			keyDoc.Set(pair.Field, order)
		}

		indexes.Append(must.NotFail(types.NewDocument(
			"pgindex", idx.PgIndex,
			"name", idx.Name,
			"key", keyDoc,
			"unique", idx.Unique,
		)))
	}

	return must.NotFail(types.NewDocument(
		"_id", c.Name,
		"table", c.TableName,
		"indexes", indexes,
	))
}

// Unmarshal sets collection metadata from [*types.Document].
func (c *Collection) Unmarshal(doc *types.Document) error {
	c.Name = must.NotFail(doc.Get("_id")).(string)
	c.TableName = must.NotFail(doc.Get("table")).(string)

	v, _ := doc.Get("indexes")
	if v == nil {
		// if there is no indexes field, nothing more to unmarshal
		return nil
	}

	arr := v.(*types.Array)

	indexes := make([]IndexInfo, arr.Len())

	for i := 0; i < arr.Len(); i++ {
		idxDoc := must.NotFail(arr.Get(i)).(*types.Document)

		idx, err := getIndexInfo(idxDoc)
		if err != nil {
			return lazyerrors.Error(err)
		}

		indexes[i] = *idx
	}

	c.Settings = Settings{Indexes: indexes}

	return nil
}

// getIndexInfo parses *types.Document to get index info.
func getIndexInfo(doc *types.Document) (*IndexInfo, error) {
	keyDoc := must.NotFail(doc.Get("key")).(*types.Document)
	key := make([]IndexKeyPair, keyDoc.Len())

	iter := keyDoc.Iterator()
	defer iter.Close()

	for i := 0; i < keyDoc.Len(); i++ {
		field, value, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		var descending bool
		switch value.(int32) {
		case int32(1):
		case int32(-1):
			descending = true
		default:
			panic(fmt.Sprintf("backends.postgresql.metadata unknown IndexKeyPair Order %v", value))
		}

		key[i] = IndexKeyPair{
			Field:      field,
			Descending: descending,
		}
	}

	// unique can be types.NullType{}
	var unique bool
	if u, ok := must.NotFail(doc.Get("unique")).(bool); ok {
		unique = u
	}

	return &IndexInfo{
		Name:    must.NotFail(doc.Get("name")).(string),
		Key:     key,
		Unique:  unique,
		PgIndex: must.NotFail(doc.Get("pgindex")).(string),
	}, nil
}
