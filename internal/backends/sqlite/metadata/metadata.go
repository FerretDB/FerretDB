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

// Collection will probably have a method for getting column name / SQLite path expression for the given document field
// once we implement field extraction.
// IDColumn probably should go away.
// TODO https://github.com/FerretDB/FerretDB/issues/226

const (
	// DefaultColumn is a column name for all fields.
	DefaultColumn = backends.ReservedPrefix + "sjson"

	// IDColumn is a SQLite path expression for _id field.
	IDColumn = DefaultColumn + "->'$._id'"
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
