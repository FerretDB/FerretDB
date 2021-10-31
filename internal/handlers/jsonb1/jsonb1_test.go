// Copyright 2021 Baltoro OÜ.
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

	"github.com/MangoDB-io/MangoDB/internal/handlers"
	"github.com/MangoDB-io/MangoDB/internal/util/testutil"
)

func setup(t testing.TB) (context.Context, *handlers.Handler, string) {
	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t)
	schema := testutil.Schema(ctx, t, pool)
	h := handlers.New(pool, zaptest.NewLogger(t), nil, nil, NewStorage(pool, zaptest.NewLogger(t)))
	return ctx, h, schema
}
