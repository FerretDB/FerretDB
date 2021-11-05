// Copyright 2021 Baltoro OÜ.
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

package jsonb1

import (
	"github.com/jackc/pgx/v4"

	"github.com/MangoDB-io/MangoDB/internal/bson"
	"github.com/MangoDB-io/MangoDB/internal/types"
	lazyerrors "github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
)

func nextRow(rows pgx.Rows) (*types.Document, error) {
	if !rows.Next() {
		err := rows.Err()
		if err != nil {
			err = lazyerrors.Error(err)
		}
		return nil, err
	}

	var b []byte
	if err := rows.Scan(&b); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var doc bson.Document
	if err := doc.UnmarshalJSON(b); err != nil {
		return nil, lazyerrors.Error(err)
	}

	d := types.MustNewDocument(&doc)
	return &d, nil
}
