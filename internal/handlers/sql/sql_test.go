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

package sql

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// All tests in this package should be fine running in parallel with tests from the handler package.
// For that reason, we use read-only access.
//
// Exported methods should be tested by the handler package.
func setup(t *testing.T) (context.Context, *pg.Pool) {
	ctx := testutil.Ctx(t)
	opts := &testutil.PoolOpts{
		ReadOnly: true,
	}
	pool := testutil.Pool(ctx, t, opts, zaptest.NewLogger(t))
	return ctx, pool
}
