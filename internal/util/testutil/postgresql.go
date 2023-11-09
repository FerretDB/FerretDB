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

package testutil

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

func CreatePostgreSQLDatabase(tb testtb.TB, ctx context.Context, p *pgxpool.Pool, name string) {
	q := fmt.Sprintf("CREATE DATABASE %s TEMPLATE template1", pgx.Identifier{name}.Sanitize())
	_, err := p.Exec(ctx, q)
	require.NoError(tb, err)
}

func DropPostgreSQLDatabase(tb testtb.TB, ctx context.Context, p *pgxpool.Pool, name string) {
	q := fmt.Sprintf("DROP DATABASE %s", pgx.Identifier{name}.Sanitize())
	_, err := p.Exec(ctx, q)
	require.NoError(tb, err)
}
