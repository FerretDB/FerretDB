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
	"fmt"
	"log/slog"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// CreateUser creates a new user and grants users with clusterAdmin role
// necessary permissions to create/update/delete users.
func CreateUser(ctx context.Context, conn *pgx.Conn, l *slog.Logger, doc *wirebson.Document) (wirebson.RawDocument, error) {
	user, _ := doc.Get("createUser").(string)
	sanitizedUser := pgx.Identifier{user}.Sanitize()

	var res wirebson.RawDocument

	err := pgx.BeginTxFunc(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error

		res, err = documentdb_api.CreateUser(ctx, tx.Conn(), l, must.NotFail(doc.Encode()))
		if err != nil {
			return lazyerrors.Error(err)
		}

		var clusterAdmin bool

		if roles := doc.Get("roles"); roles != nil {
			// valid types of "roles" is checked already by [documentdb_api.CreateUser]
			for role := range doc.Get("roles").(*wirebson.Array).Values() {
				if roleName := role.(*wirebson.Document).Get("role").(string); roleName == "clusterAdmin" {
					clusterAdmin = true

					break
				}
			}
		}

		if !clusterAdmin {
			return nil
		}

		l.DebugContext(ctx, "Updating user with CREATEROLE and ADMIN OPTION", slog.String("user", user))

		q := fmt.Sprintf("ALTER ROLE %s CREATEROLE", sanitizedUser)
		if _, err = tx.Exec(ctx, q); err != nil {
			return lazyerrors.Error(err)
		}

		// ADMIN OPTION is necessary for creating users
		q = fmt.Sprintf("GRANT documentdb_admin_role, documentdb_readonly_role TO %s WITH ADMIN OPTION", sanitizedUser)
		if _, err = tx.Exec(ctx, q); err != nil {
			return lazyerrors.Error(err)
		}

		return nil
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}
