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

import "github.com/FerretDB/FerretDB/internal/backends"

const (
	// DefaultColumn is a column name for all fields.
	DefaultColumn = backends.ReservedPrefix + "sjson"

	// IDColumn is a MySQL path expression for _id field.
	IDColumn = DefaultColumn + "->'$._id'"

	// TableIdxColumn is a column name for MySQL generated column
	TableIdxColumn = DefaultColumn + "_table"

	// RecordIDColumn is a name for RecordID column to store capped collection record id.
	RecordIDColumn = backends.ReservedPrefix + "record_id"
)

// Collection represents collection metadata.
//
// Collection value should be immutable to avoid data races.
// Use [deepCopy] to replace the whole value instead of modifying fields of existing value.
type Collection struct {
	Name            string
	TableName       string
	CappedSize      int64
	CappedDocuments int64
}

// Capped returns true if collection is capped.
func (c Collection) Capped() bool {
	return c.CappedSize > 0
}

// deepCopy returns a deep copy.
func (c *Collection) deepCopy() *Collection {
	if c == nil {
		return nil
	}

	return &Collection{
		Name:            c.Name,
		TableName:       c.TableName,
		CappedSize:      c.CappedSize,
		CappedDocuments: c.CappedDocuments,
	}
}
