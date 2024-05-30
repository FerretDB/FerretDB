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

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// DefaultColumn is a column name for all fields.
	DefaultColumn = backends.ReservedPrefix + "sjson"

	// IDColumn is a MySQL path expression for _id field.
	IDColumn = DefaultColumn + "->'$._id'"

	// IDIndexColumn is a column name for MySQL generated column on the field '_id'.
	IDIndexColumn = DefaultColumn + "_id"

	// TableIndexColumn is a column name for MySQL generated column on the field 'table'.
	TableIndexColumn = DefaultColumn + "_table"

	// RecordIDColumn is a name for RecordID column to store capped collection record id.
	RecordIDColumn = backends.ReservedPrefix + "record_id"
)

// Collection represents collection metadata.
//
// Collection value should be immutable to avoid data races.
// Use [deepCopy] to replace whole value instead of modifying fields of existing value.
type Collection struct {
	Name            string
	UUID            string
	TableName       string
	Indexes         Indexes
	CappedSize      int64
	CappedDocuments int64
}

// deepCopy returns a deep copy.
func (c *Collection) deepCopy() *Collection {
	if c == nil {
		return nil
	}

	return &Collection{
		Name:            c.Name,
		UUID:            c.UUID,
		TableName:       c.TableName,
		Indexes:         c.Indexes.deepCopy(),
		CappedSize:      c.CappedSize,
		CappedDocuments: c.CappedDocuments,
	}
}

// Capped returns true if collection is capped.
func (c Collection) Capped() bool {
	return c.CappedSize > 0
}

// Value implements driver.Valuer interface.
func (c Collection) Value() (driver.Value, error) {
	b, err := sjson.Marshal(c.marshal())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}

// Scan implements sql.Scanner interface.
func (c *Collection) Scan(src any) error {
	var doc *types.Document
	var err error

	switch src := src.(type) {
	case nil:
		*c = Collection{}
		return nil
	case []byte:
		doc, err = sjson.Unmarshal(src)
	case string:
		doc, err = sjson.Unmarshal([]byte(src))
	default:
		panic("can't scan collection")
	}

	if err != nil {
		return lazyerrors.Error(err)
	}

	if err = c.unmarshal(doc); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// marshal returns the [*types.Document] for that collection.
func (c *Collection) marshal() *types.Document {
	return must.NotFail(types.NewDocument(
		"_id", c.Name,
		"uuid", c.UUID,
		"table", c.TableName,
		"indexes", c.Indexes.marshal(),
		"cappedSize", c.CappedSize,
		"cappedDocuments", c.CappedDocuments,
	))
}

// unmarshal sets collection metadata from [*types.Document].
func (c *Collection) unmarshal(doc *types.Document) error {
	v, _ := doc.Get("_id")
	c.Name, _ = v.(string)

	if c.Name == "" {
		return lazyerrors.New("collection name is empty")
	}

	v, _ = doc.Get("table")
	c.TableName, _ = v.(string)

	if c.TableName == "" {
		return lazyerrors.New("table name is empty")
	}

	v, _ = doc.Get("indexes")
	i, _ := v.(*types.Array)

	if i == nil {
		return lazyerrors.New("indexes are empty")
	}

	if err := c.Indexes.unmarshal(i); err != nil {
		return lazyerrors.Error(err)
	}

	if v, _ := doc.Get("uuid"); v != nil {
		c.UUID = v.(string)
	}

	if v, _ := doc.Get("cappedSize"); v != nil {
		c.CappedSize = v.(int64)
	}

	if v, _ := doc.Get("cappedDocuments"); v != nil {
		c.CappedSize = v.(int64)
	}

	return nil
}

// check interfaces
var (
	_ driver.Valuer = Collection{}
	_ sql.Scanner   = (*Collection)(nil)
)
