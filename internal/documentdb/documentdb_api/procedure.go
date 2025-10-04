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

package documentdb_api

import (
	"context"
	"log/slog"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	otelsemconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/bsonhex"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// DropIndexes is a wrapper for
//
//	documentdb_api.drop_indexes(p_database_name text, p_arg documentdb_core.bson, INOUT retval documentdb_core.bson DEFAULT NULL).
//
// TODO https://github.com/documentdb/documentdb/issues/49
//
//nolint:lll // copied from generated code
func DropIndexes(ctx context.Context, conn *pgx.Conn, l *slog.Logger, databaseName string, arg wirebson.RawDocument, retVal wirebson.RawDocument) (outRetVal wirebson.RawDocument, err error) {
	ctx, span := otel.Tracer("").Start(
		ctx,
		"documentdb_api.DropIndexes",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(
			otelsemconv.DBStoredProcedureName("documentdb_api.drop_indexes"),
		),
	)
	defer span.End()

	row := conn.QueryRow(ctx, "CALL documentdb_api.drop_indexes($1, $2::bytea, $3::bytea)", databaseName, arg, retVal)

	var b []byte

	if err = row.Scan(&b); err != nil {
		err = mongoerrors.Make(ctx, err, "documentdb_api.drop_indexes", l)

		return
	}

	if outRetVal, err = bsonhex.Decode(b); err != nil {
		err = mongoerrors.Make(ctx, err, "documentdb_api.drop_indexes", l)
	}

	return
}
