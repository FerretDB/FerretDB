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

package dataapi

import (
	"testing"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb"
	"github.com/FerretDB/FerretDB/v2/internal/handler"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestSmokeDataAPI(t *testing.T) {
	const uri = "postgres://username:password@127.0.0.1:5432/postgres"
	const addr = ""

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	l := testutil.Logger(t)

	p, err := documentdb.NewPool(uri, l, sp)
	require.NoError(t, err)

	handlerOpts := &handler.NewOpts{
		Pool: p,
		Auth: true,

		L:             logging.WithName(l, "handler"),
		StateProvider: sp,
	}

	h, err := handler.New(handlerOpts)
	require.NoError(t, err)

	var lis *Listener
	lis, err = Listen(&ListenOpts{
		TCPAddr: addr,
		L:       l,
		Handler: h,
	})
	require.NoError(t, err)

	lis.Run()
}
