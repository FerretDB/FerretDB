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

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// DefaultColumn is a column name for all fields.
	DefaultColumn = "_jsonb"

	// IDColumn is a PostgreSQL path expression for _id field.
	IDColumn = DefaultColumn + "->'_id'"
)

// Collection represents collection metadata.
//
// Collection value should be immutable to avoid data races.
// Use [deepCopy] to replace the whole value instead of modifying fields of existing value.
type Collection struct {
	Name      string
	TableName string
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

// Marshal returns [*types.Document] for that collection.
func (c *Collection) Marshal() *types.Document {
	return must.NotFail(types.NewDocument(
		"_id", c.Name,
		"table", c.TableName,
		"settings", c.Settings.Marshal(),
	))
}

// Unmarshal sets collection metadata from [*types.Document].
func (c *Collection) Unmarshal(doc *types.Document) error {
	c.Name = must.NotFail(doc.Get("_id")).(string)
	c.TableName = must.NotFail(doc.Get("table")).(string)

	if doc.Has("settings") {
		var settings Settings
		must.NoError(settings.Unmarshal(must.NotFail(doc.Get("settings")).(*types.Document)))
		c.Settings = settings
	} else {
		// If settings are not present, we initialize them with empty indexes to avoid potential nil pointers.
		c.Settings = Settings{Indexes: []IndexInfo{}}
	}

	return nil
}

// Settings represents collection settings.
type Settings struct {
	Indexes []IndexInfo `json:"indexes"`
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

// Marshal returns [*types.Document] for settings.
func (s Settings) Marshal() *types.Document {
	indexes := types.MakeArray(len(s.Indexes))

	for _, index := range s.Indexes {
		key := types.MakeArray(len(index.Key))

		for _, pair := range index.Key {
			key.Append(must.NotFail(types.NewDocument(
				"field", pair.Field,
				"descending", pair.Descending,
			)))
		}

		indexes.Append(must.NotFail(types.NewDocument(
			"name", index.Name,
			"dbindex", index.DBIndex,
			"key", key,
			"unique", index.Unique,
		)))
	}

	return must.NotFail(types.NewDocument(
		"indexes", indexes,
	))
}

// Unmarshal sets settings from [*types.Document].
func (s *Settings) Unmarshal(doc *types.Document) error {
	indexes := must.NotFail(doc.Get("indexes")).(*types.Array)

	s.Indexes = make([]IndexInfo, indexes.Len())

	iter := indexes.Iterator()
	defer iter.Close()

	for {
		i, v, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		doc := v.(*types.Document)

		s.Indexes[i] = IndexInfo{
			Name:    must.NotFail(doc.Get("name")).(string),
			DBIndex: must.NotFail(doc.Get("dbindex")).(string),
			Key:     must.NotFail(doc.Get("key")).([]IndexKeyPair),
			Unique:  must.NotFail(doc.Get("unique")).(bool),
		}
	}

	return nil
}

// IndexInfo represents information about a single index.
type IndexInfo struct {
	Name    string         `json:"name"`
	DBIndex string         `json:"dbindex"` // how the index is created in the DB, like TableName for Collection
	Key     []IndexKeyPair `json:"key"`
	Unique  bool           `json:"unique"`
}

// IndexKeyPair consists of a field name and a sort order that are part of the index.
type IndexKeyPair struct {
	Field      string `json:"field"`
	Descending bool   `json:"descending"`
}
