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

package metadata

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

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

// Value implements driver.Valuer interface.
func (s Settings) Value() (driver.Value, error) {
	res, err := json.Marshal(s)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return string(res), nil
}

// Scan implements sql.Scanner interface.
func (s *Settings) Scan(src any) error {
	var err error

	switch src := src.(type) {
	case nil:
		*s = Settings{}
	case []byte:
		err = json.Unmarshal(src, s)
	case string:
		err = json.Unmarshal([]byte(src), s)
	default:
		panic("can't scan collection settings")
	}

	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// check interfaces
var (
	_ driver.Valuer = Settings{}
	_ sql.Scanner   = (*Settings)(nil)
)
