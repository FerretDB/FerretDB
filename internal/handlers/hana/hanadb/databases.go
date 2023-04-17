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

// CreateSchema creates a schema in SAP HANA JSON Document Store.
func (hanaPool *Pool) CreateSchema(ctx context.Context, db string) error {
	sqlStmt := fmt.Sprintf("CREATE SCHEMA \"%s\"", db)

	_, err := hanaPool.ExecContext(ctx, sqlStmt)

	return getHanaErrorIfExists(err)
}

// DropSchema drops database
//
// It returns ErrSchemaNotExist if schema does not exist.
func (hanaPool *Pool) DropSchema(ctx context.Context, db string) error {
	sql := fmt.Sprintf("DROP SCHEMA \"%s\" CASCADE", db)

	_, err := hanaPool.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}
