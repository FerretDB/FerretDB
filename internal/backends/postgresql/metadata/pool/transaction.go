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

package pool

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FerretDB/FerretDB/internal/util/observability"
)

// InTransaction uses pool p and wraps the given function f in a transaction.
//
// If f returns an error or context is canceled, the transaction is rolled back.
func InTransaction(ctx context.Context, p *pgxpool.Pool, f func(tx pgx.Tx) error) error {
	defer observability.FuncCall(ctx)()

	if err := pgx.BeginFunc(ctx, p, f); err != nil {
		// do not wrap error because the caller of f depends on it in some cases
		return err
	}

	return nil
}
