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

// Package mongoerrors provides MongoDB-compatible error types and codes.
package mongoerrors

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/v2/internal/util/devbuild"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

//go:generate go run ./generate.go
//go:generate ../../bin/stringer -linecomment -type Code

// Code represents MongoDB error code.
type Code int32

// Regex for missing functions/operators/etc.
var missingRE = regexp.MustCompile(
	`(function .+ does not exist|operator does not exist|operator not defined|not implemented yet)`,
)

// goString returns [fmt.GoStringer]-like string representation of the error.
func goString(err error) string {
	if err == nil {
		return "<nil>"
	}

	switch err := err.(type) { //nolint:errorlint // we intentionally do not inspect the error chain
	case *pgconn.ParseConfigError:
		// %#v would produce `err:(*errors.errorString)(0xc000237550)`
		return fmt.Sprintf("&pgconn.ParseConfigError(%q)", err.Error())
	case *pgconn.ConnectError:
		// %#v would produce
		// &pgconn.ConnectError{Config: (*pgconn.Config)(0x1400031e360), err: (*fmt.wrapError)(0x1400030a160)}
		return fmt.Sprintf("&pgconn.ConnectError(%q)", err.Error())
	case *pgconn.PgError:
		return fmt.Sprintf("%#v", err)
	default:
		return fmt.Sprintf("%#v", err)
	}
}

// Make converts any error to [*Error].
//
// Nil panics (it never should be passed),
// [*Error] (possibly wrapped) is returned unwrapped,
// [*pgconn.PgError] (possibly wrapped) is converted by mapping error code,
// any other values are returned as [*Error] with [ErrInternalError] code.
//
// It also records error to the current Otel span.
//
// This function performs double duty: it is used to convert errors in documentdb_api,
// and to map error codes in conn.go. It probably should be split in two.
func Make(ctx context.Context, err error, arg string, l *slog.Logger) *Error {
	must.NotBeZero(err)

	span := oteltrace.SpanFromContext(ctx)
	span.SetStatus(otelcodes.Error, "")
	span.RecordError(err)

	var e *Error
	if errors.As(err, &e) {
		return e
	}

	var pg *pgconn.PgError
	if !errors.As(err, &pg) {
		l.WarnContext(ctx, "Unexpected error type", slog.String("arg", arg), slog.String("error", goString(err)))

		return &Error{
			Argument: arg,
			CommandError: mongo.CommandError{
				Code:    int32(ErrInternalError),
				Message: err.Error(),
				Name:    ErrInternalError.String(),
				Wrapped: err,
			},
		}
	}

	if devbuild.Enabled && missingRE.MatchString(pg.Message) {
		l.LogAttrs(ctx, logging.LevelDPanic, "Missing", slog.String("arg", arg), slog.String("error", goString(err)))
	}

	var code Code

	switch pg.Code {
	case pgerrcode.UndefinedFunction:
		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/853
		l.ErrorContext(ctx, "Missing function", slog.String("arg", arg), slog.String("error", goString(err)))
		code = ErrInternalError

	case pgerrcode.ConnectionFailure:
		// mainly for tests
		l.ErrorContext(ctx, "Connection failure", slog.String("arg", arg), slog.String("error", goString(err)))
		code = ErrInternalError
	}

	if len(pg.Code) == 5 && pg.Code[0] == 'M' {
		code = pgCodes[pg.Code]
	}

	if code == 0 {
		level := logging.LevelDPanic

		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/894
		if arg == "documentdb_api.rename_collection" || arg == "documentdb_api.find_and_modify" {
			level = slog.LevelError
		}

		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1147
		if arg == "documentdb_api_internal.create_indexes_non_concurrently" {
			level = slog.LevelError
		}

		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/914
		if arg == "documentdb_api.create_user" || arg == "documentdb_api.update_user" || arg == "documentdb_api.drop_user" {
			level = slog.LevelError
		}

		l.LogAttrs(ctx, level, "Unmapped error code", slog.String("arg", arg), slog.String("error", goString(err)))
		code = ErrInternalError
	}

	return &Error{
		Argument: arg,
		CommandError: mongo.CommandError{
			Code:    int32(code),
			Message: pg.Message,
			Name:    code.String(),
			Wrapped: err,
		},
	}
}

// MapWrappedCode maps error code found inside "writeErrors" responses for insert/update/delete operations
// and inside createIndexes responses.
//
// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/292
// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/895
func MapWrappedCode(code int32) Code {
	switch code {
	case 16777245:
		return ErrBadValue // 2
	case 50331677:
		return ErrFailedToParse // 9
	case 67108893:
		return ErrTypeMismatch // 14
	case 285212701:
		return ErrPathNotViable // 28
	case 319029277:
		return ErrDuplicateKey // 11000
	case 335544349:
		return ErrConflictingUpdateOperators // 40
	case 385875997:
		return ErrDollarPrefixedFieldName // 52
	case 436207645:
		return ErrEmptyFieldName // 56
	case 486539293:
		return ErrImmutableField // 66
	case 503316509:
		return ErrCannotCreateIndex // 67
	case 520093725:
		return ErrIndexAlreadyExists // 68
	case 553648157:
		return ErrInvalidNamespace // 73
	case 570425373:
		return ErrIndexOptionsConflict // 85
	case 587202589:
		return ErrIndexKeySpecsConflict // 86
	default:
		return Code(code)
	}
}

// MapWriteErrors replaces error codes inside "writeErrors" responses.
//
// The whole function is temporary workaround for:
// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/895
func MapWriteErrors(ctx context.Context, res wirebson.AnyDocument) wirebson.AnyDocument {
	_, span := otel.Tracer("").Start(ctx, "MapWriteErrors")
	defer span.End()

	resDoc := must.NotFail(res.Decode())

	v := resDoc.Get("writeErrors")
	if v == nil {
		return res
	}

	writeErrors := must.NotFail(v.(wirebson.AnyArray).Decode())

	if writeErrors.Len() == 0 {
		return res
	}

	for i, el := range writeErrors.All() {
		writeError := must.NotFail(el.(wirebson.AnyDocument).Decode())

		code := MapWrappedCode(writeError.Get("code").(int32))

		must.NoError(writeError.Replace("code", int32(code)))
		must.NoError(writeErrors.Replace(i, writeError))
	}

	must.NoError(resDoc.Replace("writeErrors", writeErrors))

	span.SetStatus(otelcodes.Error, "")

	return resDoc
}
