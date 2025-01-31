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

package mongoerrors

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestCode(t *testing.T) {
	// conn.route depends on non-empty strings
	assert.NotEmpty(t, Code(0).String())
	assert.NotEmpty(t, Code(1).String())
}

func TestMake(t *testing.T) {
	ctx := testutil.Ctx(t)
	l := testutil.Logger(t)

	pg := &pgconn.PgError{
		Severity:            "ERROR",
		SeverityUnlocalized: "ERROR",
		Code:                "M0001",
		Message:             "unknown top level operator: $comment.",
		File:                "query_operator.c",
		Line:                1193,
		Routine:             "CreateBoolExprFromLogicalExpression",
	}

	expected := &Error{
		Argument: "documentdb_api.find_cursor_first_page",
		CommandError: mongo.CommandError{
			Code:    int32(ErrBadValue),
			Message: "unknown top level operator: $comment.",
			Name:    ErrBadValue.String(),
			Wrapped: pg,
		},
	}

	err := Make(ctx, pg, "documentdb_api.find_cursor_first_page", l)
	assert.Equal(t, expected, err)
	assert.ErrorIs(t, err, pg)

	expectedS := "BadValue (2): unknown top level operator: $comment."
	assert.Equal(t, expectedS, fmt.Sprintf("%s", err))
	assert.Equal(t, expectedS, fmt.Sprintf("%v", err))
	assert.Equal(t, expectedS, fmt.Sprintf("%+v", err))

	expectedS = "&mongoerrors.Error{" +
		"Code: 2, Name: `BadValue`, " +
		"Message: `unknown top level operator: $comment.`, " +
		"Argument: `documentdb_api.find_cursor_first_page`, " +
		"Wrapped: &pgconn.PgError{" +
		`Severity:"ERROR", SeverityUnlocalized:"ERROR", Code:"M0001", ` +
		`Message:"unknown top level operator: $comment.", Detail:"", Hint:"", Position:0, InternalPosition:0, ` +
		`InternalQuery:"", Where:"", SchemaName:"", TableName:"", ColumnName:"", DataTypeName:"", ConstraintName:"", ` +
		`File:"query_operator.c", Line:1193, Routine:"CreateBoolExprFromLogicalExpression"` +
		"}}"
	assert.Equal(t, expectedS, fmt.Sprintf("%#v", err))
}

func TestMakeParseConfigError(t *testing.T) {
	ctx := testutil.Ctx(t)
	l := testutil.Logger(t)

	_, pg := pgx.Connect(ctx, "invalid")
	assert.IsType(t, (*pgconn.ParseConfigError)(nil), pg)

	expected := &Error{
		CommandError: mongo.CommandError{
			Code:    int32(ErrInternalError),
			Message: "cannot parse `invalid`: failed to parse as keyword/value (invalid keyword/value)",
			Name:    ErrInternalError.String(),
			Wrapped: pg,
		},
	}

	err := Make(ctx, pg, "", l)
	assert.Equal(t, expected, err)
	assert.ErrorIs(t, err, pg)

	expectedS := "InternalError (1): cannot parse `invalid`: failed to parse as keyword/value (invalid keyword/value)"
	assert.Equal(t, expectedS, fmt.Sprintf("%s", err))
	assert.Equal(t, expectedS, fmt.Sprintf("%v", err))
	assert.Equal(t, expectedS, fmt.Sprintf("%+v", err))

	expectedS = "&mongoerrors.Error{" +
		"Code: 1, Name: `InternalError`, " +
		"Message: \"cannot parse `invalid`: failed to parse as keyword/value (invalid keyword/value)\", " +
		"Argument: ``, " +
		"Wrapped: &pgconn.ParseConfigError(" +
		"\"cannot parse `invalid`: failed to parse as keyword/value (invalid keyword/value)\"" +
		")}"
	assert.Equal(t, expectedS, fmt.Sprintf("%#v", err))
}

func TestMakeConnectError(t *testing.T) {
	ctx := testutil.Ctx(t)
	l := testutil.Logger(t)

	_, pg := pgx.Connect(ctx, "postgres://invalid/")
	assert.IsType(t, (*pgconn.ConnectError)(nil), pg)

	expected := &Error{
		CommandError: mongo.CommandError{
			Code:    int32(ErrInternalError),
			Name:    ErrInternalError.String(),
			Wrapped: pg,
		},
	}

	err := Make(ctx, pg, "", l)

	assert.Regexp(t, `failed to connect`, err.Message)
	expected.CommandError.Message = err.Message

	assert.Equal(t, expected, err)
	assert.ErrorIs(t, err, pg)

	expectedS := "InternalError (1): " + err.Message
	assert.Equal(t, expectedS, fmt.Sprintf("%s", err))
	assert.Equal(t, expectedS, fmt.Sprintf("%v", err))
	assert.Equal(t, expectedS, fmt.Sprintf("%+v", err))

	expectedS = "&mongoerrors.Error{" +
		"Code: 1, Name: `InternalError`, " +
		"Message: " + strconv.Quote(err.Message) + ", " +
		"Argument: ``, " +
		"Wrapped: &pgconn.ConnectError(" + strconv.Quote(err.Message) + ")}"
	assert.Equal(t, expectedS, fmt.Sprintf("%#v", err))
}
