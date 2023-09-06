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
	"encoding/json"
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
	Field      string `json:"field"`
	Descending bool   `json:"descending"`
}

// Value implements driver.Valuer interface.
func (s Settings) Value() (driver.Value, error) {
	res, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return string(res), nil
}

// Scan implements sql.Scanner interface.
func (s *Settings) Scan(src any) error {
	switch src := src.(type) {
	case nil:
		*s = Settings{}
	case []byte:
		return json.Unmarshal(src, s)
	case string:
		return json.Unmarshal([]byte(src), s)
	default:
		panic("can't scan collection settings")
	}

	return nil
}

// check interfaces
var (
	_ driver.Valuer = Settings{}
	_ sql.Scanner   = (*Settings)(nil)
)
