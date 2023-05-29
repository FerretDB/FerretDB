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
	"errors"
	"fmt"
)

// CreateSchema creates a schema in SAP HANA JSON Document Store.
//
//	Returns ErrSchemaAlreadyExist if schema already exists.
func (hanaPool *Pool) CreateSchema(ctx context.Context, qp *QueryParams) error {
	sqlStmt := fmt.Sprintf("CREATE SCHEMA %q", qp.DB)

	_, err := hanaPool.ExecContext(ctx, sqlStmt)

	return getHanaErrorIfExists(err)
}

// CreateSchemaIfNotExists creates a schema in SAP HANA JSON Document Store.
//
// Returns nil if the schema already exists.
func (hanaPool *Pool) CreateSchemaIfNotExists(ctx context.Context, qp *QueryParams) error {
	err := hanaPool.CreateSchema(ctx, qp)

	switch {
	case errors.Is(err, ErrSchemaAlreadyExist):
		return nil
	default:
		return err
	}
}

// DropSchema drops database.
//
// Returns ErrSchemaNotExist if schema does not exist.
func (hanaPool *Pool) DropSchema(ctx context.Context, qp *QueryParams) error {
	sql := fmt.Sprintf("DROP SCHEMA %q CASCADE", qp.DB)

	_, err := hanaPool.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}
