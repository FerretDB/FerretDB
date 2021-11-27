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
	pool := testutil.Pool(ctx, tb)
	schema := testutil.Schema(ctx, tb, pool)
	h := handlers.New(pool, zaptest.NewLogger(tb), nil, nil, NewStorage(pool, zaptest.NewLogger(tb)))
	return ctx, h, schema
}
