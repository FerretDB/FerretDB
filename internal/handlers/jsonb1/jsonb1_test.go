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

package jsonb1

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/handlers"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func setup(tb testing.TB) (context.Context, *handlers.Handler, string) {
	ctx := testutil.Ctx(tb)
	pool := testutil.Pool(ctx, tb, new(testutil.PoolOpts))
	db := testutil.Schema(ctx, tb, pool)
	handlerOpts := &handlers.NewOpts{
		PgPool:        pool,
		Logger:        zaptest.NewLogger(tb),
		SharedHandler: nil,
		SQLStorage:    nil,
		JSONB1Storage: NewStorage(pool, zaptest.NewLogger(tb)),
		Metrics:       handlers.NewMetrics(),
	}
	h := handlers.New(handlerOpts)
	return ctx, h, db
}
