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

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
)

// CreateUser creates a new user.
// Users with the `clusterAdmin` role are given PostgreSQL's SUPERUSER privileges.
func CreateUser(ctx context.Context, conn *pgx.Conn, l *slog.Logger, docV wirebson.AnyDocument) (wirebson.RawDocument, error) {
	spec, err := docV.Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := docV.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res wirebson.RawDocument

	err = pgx.BeginTxFunc(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		res, err = documentdb_api.CreateUser(ctx, tx.Conn(), l, spec)
		if err != nil {
			return lazyerrors.Error(err)
		}

		var clusterAdmin bool

		if rolesV := doc.Get("roles"); rolesV != nil {
			// valid value of "roles" is checked already by [documentdb_api.CreateUser]
			var roles *wirebson.Array

			if roles, err = rolesV.(wirebson.AnyArray).Decode(); err != nil {
				return lazyerrors.Error(err)
			}

			for roleV := range roles.Values() {
				var role *wirebson.Document

				if role, err = roleV.(wirebson.AnyDocument).Decode(); err != nil {
					return lazyerrors.Error(err)
				}

				if roleName := role.Get("role").(string); roleName == "clusterAdmin" {
					clusterAdmin = true

					break
				}
			}
		}

		if !clusterAdmin {
			return nil
		}

		user, _ := doc.Get(doc.Command()).(string)
		l.DebugContext(ctx, "Updating user to SUPERUSER", slog.String("user", user))

		q := fmt.Sprintf("ALTER ROLE %s SUPERUSER", pgx.Identifier{user}.Sanitize())
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
