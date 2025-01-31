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

// Package documentdb_api_internal is generated and then reduced to what we need manually.
//
// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1148
//
//nolint:lll,wsl // generated code is not great for linters
package documentdb_api_internal

import (
	"context"
	"log/slog"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// CreateIndexesNonConcurrently is a wrapper for
//
//	documentdb_api_internal.create_indexes_non_concurrently(p_database_name text, p_arg documentdb_core.bson, p_skip_check_collection_create boolean DEFAULT false, OUT create_indexes_non_concurrently documentdb_core.bson).
//
// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1147
func CreateIndexesNonConcurrently(ctx context.Context, conn *pgx.Conn, l *slog.Logger, databaseName string, arg wirebson.RawDocument, skipCheckCollectionCreate bool) (outCreateIndexesNonConcurrently wirebson.RawDocument, err error) {
	ctx, span := otel.Tracer("").Start(ctx, "documentdb_api_internal.create_indexes_non_concurrently", oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	row := conn.QueryRow(ctx, "SELECT create_indexes_non_concurrently::bytea FROM documentdb_api_internal.create_indexes_non_concurrently($1, $2::bytea, $3)", databaseName, arg, skipCheckCollectionCreate)
	if err = row.Scan(&outCreateIndexesNonConcurrently); err != nil {
		err = mongoerrors.Make(ctx, err, "documentdb_api_internal.create_indexes_non_concurrently", l)
	}
	return
}

// ScramSha256GetSaltAndIterations is a wrapper for
//
//	documentdb_api_internal.scram_sha256_get_salt_and_iterations(p_user_name text, OUT scram_sha256_get_salt_and_iterations documentdb_core.bson).
func ScramSha256GetSaltAndIterations(ctx context.Context, conn *pgx.Conn, l *slog.Logger, userName string) (outScramSha256GetSaltAndIterations wirebson.RawDocument, err error) {
	ctx, span := otel.Tracer("").Start(ctx, "documentdb_api_internal.scram_sha256_get_salt_and_iterations", oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	row := conn.QueryRow(ctx, "SELECT scram_sha256_get_salt_and_iterations::bytea FROM documentdb_api_internal.scram_sha256_get_salt_and_iterations($1)", userName)
	if err = row.Scan(&outScramSha256GetSaltAndIterations); err != nil {
		err = mongoerrors.Make(ctx, err, "documentdb_api_internal.scram_sha256_get_salt_and_iterations", l)
	}
	return
}

// AuthenticateWithScramSha256 is a wrapper for
//
//	documentdb_api_internal.authenticate_with_scram_sha256(p_user_name text, p_auth_msg text, p_client_proof text, OUT authenticate_with_scram_sha256 documentdb_core.bson).
func AuthenticateWithScramSha256(ctx context.Context, conn *pgx.Conn, l *slog.Logger, userName string, authMsg string, clientProof string) (outAuthenticateWithScramSha256 wirebson.RawDocument, err error) {
	ctx, span := otel.Tracer("").Start(ctx, "documentdb_api_internal.authenticate_with_scram_sha256", oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	defer span.End()

	row := conn.QueryRow(ctx, "SELECT authenticate_with_scram_sha256::bytea FROM documentdb_api_internal.authenticate_with_scram_sha256($1, $2, $3)", userName, authMsg, clientProof)
	if err = row.Scan(&outAuthenticateWithScramSha256); err != nil {
		err = mongoerrors.Make(ctx, err, "documentdb_api_internal.authenticate_with_scram_sha256", l)
	}
	return
}
