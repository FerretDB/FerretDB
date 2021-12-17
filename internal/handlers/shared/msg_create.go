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

package shared

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
	"github.com/jackc/pgx/v4"
)

// CreateCollection creates a new collection in db.
func CreateCollection(ctx context.Context, pgPool *pg.Pool, db, collection string) error {
	sql := `CREATE TABLE ` + pgx.Identifier{db, collection}.Sanitize() + ` (_jsonb jsonb)`
	_, err := pgPool.Exec(ctx, sql)
	return err
}

// MsgCreate adds a collection or view into the database.
func (h *Handler) MsgCreate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	m := document.Map()
	collection := m[document.Command()].(string)
	db := m["$db"].(string)

	if err := CreateCollection(ctx, h.pgPool, db, collection); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"created", db+"."+collection,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
