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

package register

import (
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/handlers/pg"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
)

// InitPg registers `pg` handler for PostgreSQL that is always enabled.
func InitPg(conn string) {
	RegisteredHandlers["pg"] = func(opts *NewHandlerOpts) (handlers.Interface, error) {
		pgPool, err := pgdb.NewPool(opts.Ctx, opts.PostgreSQLConnectionString, opts.Logger, false)
		if err != nil {
			return nil, err
		}

		handlerOpts := &pg.NewOpts{
			PgPool:    pgPool,
			L:         opts.Logger,
			StartTime: time.Now(),
		}
		return pg.New(handlerOpts)
	}
}
