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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// InsertOne inserts a document into a collection in SAP HANA JSON Document Store.
func (hanaPool *Pool) InsertOne(ctx context.Context, qp *QueryParams, doc *types.Document) error {
	err := hanaPool.CreateSchemaIfNotExists(ctx, qp)

	switch {
	case err == nil:
		// Success case
	default:
		return err
	}

	err = hanaPool.CreateCollectionIfNotExists(ctx, qp)

	switch {
	case err == nil:
		// Success case
	default:
		return err
	}

	return hanaPool.insert(ctx, qp, doc)
}

// insert inserts a document into a collection in SAP HANA JSON Document Store.
func (hanaPool *Pool) insert(ctx context.Context, qp *QueryParams, doc *types.Document) error {
	sqlStmt := fmt.Sprintf("insert into %q.%q values($1)", qp.DB, qp.Collection)

	_, err := hanaPool.ExecContext(ctx, sqlStmt, must.NotFail(Marshal(doc)))

	return getHanaErrorIfExists(err)
}
