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
	"slices"

	"github.com/FerretDB/FerretDB/internal/types"
)

// Collection represents collection metadata.
type Collection struct {
	Name      string `json:"_id"`
	TableName string `json:"table"`
	Settings
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
	Name   string         `json:"name"`
	Key    []IndexKeyPair `json:"key"`
	Unique bool           `json:"unique"`
}

// IndexKeyPair consists of a field name and a sort order that are part of the index.
type IndexKeyPair struct {
	// TODO
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
	panic("not implemented")
}

// Unmarshal sets collection metadata from [*types.Document].
func (c *Collection) Unmarshal(doc *types.Document) error {
	panic("not implemented")
}
