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

package documentdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// CreateUser creates a new user and grants it the necessary role and permissions
// to create/update/delete users.
// It uses a transaction to ensure that roles and permissions are set atomically.
func CreateUser(ctx context.Context, conn *pgx.Conn, l *slog.Logger, doc *wirebson.Document) (res wirebson.RawDocument, err error) { //nolint:lll // for readability
	user, _ := doc.Get("createUser").(string)
	sanitizedUser := pgx.Identifier{user}.Sanitize()

	var tx pgx.Tx

	if tx, err = conn.Begin(ctx); err != nil {
		err = lazyerrors.Error(err)

		return
	}

	defer func() {
		if e := tx.Rollback(ctx); e != nil && !errors.Is(e, pgx.ErrTxClosed) && err == nil {
			err = lazyerrors.Error(e)
		}
	}()

	res, err = documentdb_api.CreateUser(ctx, tx.Conn(), l, must.NotFail(doc.Encode()))
	if err != nil {
		err = lazyerrors.Error(err)

		return
	}

	q := fmt.Sprintf("ALTER ROLE %s CREATEROLE", sanitizedUser)
	if _, err = tx.Exec(ctx, q); err != nil {
		err = lazyerrors.Error(err)

		return
	}

	// ADMIN OPTION is necessary for creating users
	q = fmt.Sprintf("GRANT documentdb_admin_role, documentdb_readonly_role TO %s WITH ADMIN OPTION", sanitizedUser)
	if _, err = tx.Exec(ctx, q); err != nil {
		err = lazyerrors.Error(err)

		return
	}

	if err = tx.Commit(ctx); err != nil {
		err = lazyerrors.Error(err)

		return
	}

	return
}
