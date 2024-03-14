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

package driver

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestDriver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	ctx := testutil.Ctx(t)

	c, err := Connect(ctx, "mongodb://127.0.0.1:47017/", testutil.SLogger(t))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, c.Close()) })

	c.Write()

	// TODO https://github.com/FerretDB/FerretDB/issues/4146
	_ = c
}
