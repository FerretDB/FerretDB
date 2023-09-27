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

package integration

import (
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// resultPushdown stores the information about expected pushdown results for a single or multiple backends.
// For example if both pg and SQlite backends are expected to pushdown, the `pgPushdown + sqlitePushdown` operation
// can be used.
type resultPushdown uint8

const (
	// List of supported backends.
	pgPushdown     resultPushdown = 1 << iota // 0000 0001
	sqlitePushdown                            // 0000 0010

	// No pushdown expected.
	noPushdown resultPushdown = 0 // 0000 0000

	// Expects all backends to result in pushdown.
	allPushdown resultPushdown = 0xff // 1111 1111
)

func init() {
	must.BeTrue(noPushdown == 0b00000000)
	must.BeTrue(pgPushdown == 0b00000001)
	must.BeTrue(sqlitePushdown == 0b00000010)
	must.BeTrue(allPushdown == 0b11111111)
}

// PushdownExpected returns true if the pushdown is expected for currently running backend.
func (res resultPushdown) PushdownExpected(t testtb.TB) bool {
	if setup.IsPushdownDisabled() {
		res = noPushdown
	}

	switch {
	case setup.IsPostgres(t):
		return res&pgPushdown == pgPushdown
	case setup.IsSQLite(t):
		return res&sqlitePushdown == sqlitePushdown
	case setup.IsMongoDB(t):
		return false
	default:
		panic("Unknown backend")
	}
}
