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

// Package metadata provides access to SQLite databases and collections information.
package metadata

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
	Settings  []byte
}
