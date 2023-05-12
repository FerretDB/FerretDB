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

package backend

type Database interface {
	CreateCollection(params *CreateCollectionParams) error
	DropCollection(params *DropCollectionParams) error
}

func DatabaseContract(db Database) Database {
	return &databaseContract{
		db: db,
	}
}

type databaseContract struct {
	db Database
}

type CreateCollectionParams struct{}

func (db *databaseContract) CreateCollection(params *CreateCollectionParams) (err error) {
	defer checkError(err, ErrCollectionAlreadyExists)
	err = db.db.CreateCollection(params)
	return
}

type DropCollectionParams struct{}

func (db *databaseContract) DropCollection(params *DropCollectionParams) (err error) {
	defer checkError(err, ErrCollectionAlreadyExists)
	err = db.db.DropCollection(params)
	return
}

// check interfaces
var (
	_ Database = (*databaseContract)(nil)
)
