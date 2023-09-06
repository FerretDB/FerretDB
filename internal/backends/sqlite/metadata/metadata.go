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
	"database/sql"
	"database/sql/driver"
)

// Collection will probably have a method for getting column name / SQLite path expression for the given document field
// once we implement field extraction.
// IDColumn probably should go away.
// TODO https://github.com/FerretDB/FerretDB/issues/226

const (
	// IDColumn is a SQLite path expression for _id field.
	IDColumn = "_ferretdb_sjson->'$._id'"

	// DefaultColumn is a column name for all fields expect _id.
	DefaultColumn = "_ferretdb_sjson"
)

// Collection represents collection metadata.
type Collection struct {
	Name      string
	TableName string
	Settings  Settings
}

type Settings struct {
	Indexes []IndexInfo
}

func (s Settings) Value() (driver.Value, error) {

}

func (s *Settings) Scan(...) error {

}

var _ driver.Valuer = Settings{}
var _ sql.Scanner = (*Settings)(nil)

// CollectionSettings represents collection settings model.
type CollectionSettings struct {
	Indexes []IndexInfo
}

// IndexInfo represents information about a single index.
type IndexInfo struct {
	Name   string
	Key    []IndexKeyPair
	Unique bool
}

// IndexKeyPair consists of a field name and a sort order that are part of the index.
type IndexKeyPair struct {
	Field      string
	Descending bool
}
