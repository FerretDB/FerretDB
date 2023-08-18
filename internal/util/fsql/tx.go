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

package fsql

import (
	"database/sql"

	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// Tz wraps [*database/sql.Tx] with resource tracking.
//
// It exposes the subset of *sql.Tx methods we use.
type Tx struct {
	sqlTx *sql.Tx
	token *resource.Token
}

// wrapTx creates new Tx.
func wrapTx(tx *sql.Tx) *Tx {
	if tx == nil {
		return nil
	}

	res := &Tx{
		sqlTx: tx,
		token: resource.NewToken(),
	}

	resource.Track(res, res.token)

	return res
}

// Commit calls [*sql.Tx.Commit].
func (tx *Tx) Commit() error {
	resource.Untrack(tx, tx.token)
	return tx.sqlTx.Commit()
}

// Rollback calls [*sql.Tx.Rollback].
func (tx *Tx) Rollback() error {
	resource.Untrack(tx, tx.token)
	return tx.sqlTx.Rollback()
}
