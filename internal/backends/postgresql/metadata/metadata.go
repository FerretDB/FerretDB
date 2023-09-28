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
type Collection struct {
	Name      string
	TableName string
	// TODO https://github.com/FerretDB/FerretDB/issues/3375
}

// deepCopy returns a deep copy.
func (c *Collection) deepCopy() *Collection {
	if c == nil {
		return nil
	}

	return &Collection{
		Name:      c.Name,
		TableName: c.TableName,
	}
}

// Marshal returns [*types.Document] for that collection.
func (c *Collection) Marshal() *types.Document {
	return must.NotFail(types.NewDocument(
		"_id", c.Name,
		"table", c.TableName,
	))
}

// Unmarshal sets collection metadata from [*types.Document].
func (c *Collection) Unmarshal(doc *types.Document) error {
	c.Name = must.NotFail(doc.Get("_id")).(string)
	c.TableName = must.NotFail(doc.Get("table")).(string)

	return nil
}
