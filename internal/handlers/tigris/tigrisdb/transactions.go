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

package tigrisdb

import (
	"context"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

func (tdb *TigrisDB) InTransaction(ctx context.Context, f func(driver.Tx) error, db string) (err error) {
	var tx driver.Tx

	if tx, err = tdb.Driver.UseDatabase(db).BeginTx(ctx, nil); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	defer func() {
		if err == nil {
			return
		}

		if rerr := tx.Rollback(ctx); rerr != nil {
			//	tdb.l.Error("failed to perform rollback", "error", rerr)
		}
	}()

	if err = f(tx); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	if err = tx.Commit(ctx); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	return
}
