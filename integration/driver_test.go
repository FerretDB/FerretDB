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
	"testing"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestDriver(t *testing.T) {
	t.Parallel()

	ctx, conn := setup.SetupWireConn(t)

	t.Run("Drop", func(t *testing.T) {
		dropCmd := must.NotFail(wirebson.NewDocument(
			"dropDatabase", int32(1),
			"$db", testutil.DatabaseName(t),
		))

		_, resBody, err := conn.Request(ctx, must.NotFail(wire.NewOpMsg(dropCmd)))
		require.NoError(t, err)

		resMsg, err := must.NotFail(resBody.(*wire.OpMsg).RawDocument()).Decode()
		require.NoError(t, err)

		ok := resMsg.Get("ok").(float64)
		require.Equal(t, float64(1), ok)
	})
}
