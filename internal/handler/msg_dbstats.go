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

package handler

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
)

// msgDBStats implements `dbStats` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgDBStats(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	var res wirebson.RawDocument

	err = h.p.WithConnContext(connCtx, func(conn *pgx.Conn) error {
		res, err = documentdb_api.DbStats(connCtx, conn, h.L, dbName, 1, true)
		if err != nil || os.Getenv("FERRETDB_UNIFI_COMPAT") != "true" {
			return err
		}
		// Native relation statistics lag committed writes. Compute logical BSON
		// totals from one statement snapshot; retain native physical sizes.
		rows, e := conn.Query(connCtx, "SELECT collection_id FROM documentdb_api_catalog.collections WHERE database_name=$1 AND view_definition IS NULL AND shard_key IS NULL", dbName)
		if e != nil {
			return e
		}
		var parts []string
		for rows.Next() {
			var id int64
			if e = rows.Scan(&id); e != nil {
				rows.Close()
				return e
			}
			table := pgx.Identifier{"documentdb_data", fmt.Sprintf("documents_%d", id)}.Sanitize()
			parts = append(parts, "SELECT count(*)::numeric AS objects, COALESCE(sum(octet_length(documentdb_core.bson_to_bytea(document))),0)::numeric AS bytes FROM "+table)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		var count, size int64
		if len(parts) > 0 {
			e = conn.QueryRow(connCtx, "SELECT sum(objects)::bigint,sum(bytes)::bigint FROM ("+strings.Join(parts, " UNION ALL ")+") AS totals").Scan(&count, &size)
			if e != nil {
				return e
			}
		}
		decoded, e := res.Decode()
		if e != nil {
			return e
		}
		decoded = decoded.Copy()
		average := float64(0)
		if count > 0 {
			average = float64(size) / float64(count)
		}
		for key, value := range map[string]any{"objects": count, "dataSize": float64(size), "avgObjSize": average} {
			if e = decoded.Replace(key, value); e != nil {
				return e
			}
		}
		res, e = decoded.Encode()
		return e
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return middleware.ResponseDoc(req, res)
}
