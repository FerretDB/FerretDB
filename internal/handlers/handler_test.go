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

package handlers

import (
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/MangoDB-io/MangoDB/internal/handlers/jsonb1"
	"github.com/MangoDB-io/MangoDB/internal/handlers/shared"
	"github.com/MangoDB-io/MangoDB/internal/handlers/sql"
	"github.com/MangoDB-io/MangoDB/internal/util/testutil"
)

func TestCRUD(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := testutil.Pool(ctx, t)
	schema := testutil.Schema(ctx, t, pool)
	l := zaptest.NewLogger(t)
	shared := shared.NewHandler(pool, "127.0.0.1:12345")
	sql := sql.NewStorage(pool, l.Sugar())
	jsonb1 := jsonb1.NewStorage(pool, l)
	handler := New(pool, l, shared, sql, jsonb1)

	// handler.Handle(ctx, &wire.MsgHeader{}, nil)
	_ = schema
	_ = handler
}
