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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// QueryParams represents options/parameters used for SQL query/statement.
type QueryParams struct {
	DB         string
	Collection string
}

func (hanaPool *Pool) QueryDocuments(ctx context.Context, qp *QueryParams) ([]*types.Document, error) {

	sqlStmt := fmt.Sprintf("SELECT %q FROM %q", qp.Collection, qp.DB)

	rows, err := hanaPool.QueryContext(ctx, sqlStmt)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var documents []*types.Document

	// Todo: transform rows into documents
	for rows.Next() {
		var docStr string
		if err = rows.Scan(&docStr); err != nil {
			return nil, lazyerrors.Error(err)
		}
		// Todo: create document from rowString
		//documents = append(documents, types.NewDocument(docStr))
	}

	return documents, err
}
