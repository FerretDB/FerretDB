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
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

func TestMsgDropUserNotFoundMapping(t *testing.T) {
	t.Parallel()

	user := "non-existent-user"
	dbName := "test"

	pgErr := &pgconn.PgError{
		Code:    pgerrcode.UndefinedObject,
		Message: `role "non-existent-user" does not exist`,
	}

	mErr := mongoerrors.NewWithArgument(
		mongoerrors.ErrUserNotFound,
		pgErr.Message,
		"documentdb_api.drop_user",
	)

	for _, origErr := range []error{pgErr, mErr} {
		var targetMErr *mongoerrors.Error
		var targetPgErr *pgconn.PgError

		isUserNotFound := (errors.As(origErr, &targetMErr) && targetMErr.Code == int32(mongoerrors.ErrUserNotFound)) ||
			(errors.As(origErr, &targetPgErr) && targetPgErr.Code == pgerrcode.UndefinedObject)

		require.True(t, isUserNotFound)

		resErr := mongoerrors.NewWithArgument(
			mongoerrors.ErrUserNotFound,
			fmt.Sprintf("User '%s@%s' not found", user, dbName),
			"dropUser",
		)

		assert.Equal(t, int32(mongoerrors.ErrUserNotFound), resErr.Code)
		assert.Equal(t, "UserNotFound", resErr.Name)
		assert.Equal(t, "User 'non-existent-user@test' not found", resErr.Message)
		assert.Equal(t, "dropUser", resErr.Argument)
	}
}
