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

package hanadb

import (
	"context"
	"fmt"

	_ "github.com/SAP/go-hdb/driver"
)

// CreateCollection creates a new SAP HANA JSON Document Store collection.
//
// It returns ErrAlreadyExist if collection already exist.
func (hanaPool *Pool) CreateCollection(ctx context.Context, db, collection string) error {
	sql := fmt.Sprintf("CREATE COLLECTION \"%s\".\"%s\"", db, collection)

	_, err := hanaPool.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

// DropCollection drops collection
//
// It returns ErrTableNotExist is collection does not exist.
func (hanaPool *Pool) DropCollection(ctx context.Context, db, collection string) error {
	sql := fmt.Sprintf("DROP COLLECTION \"%s\".\"%s\"", db, collection)

	_, err := hanaPool.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}
