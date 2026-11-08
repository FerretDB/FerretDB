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

package handler

import (
	"testing"
	"time"

	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/handler/session"
	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestMsgKillCursorsDecodeError(t *testing.T) {
	t.Parallel()

	connInfo := conninfo.New()
	defer connInfo.Close()

	connCtx := conninfo.Ctx(t.Context(), connInfo)

	s := session.NewRegistry(30*time.Minute, testutil.Logger(t))
	defer s.Stop()

	h := &Handler{
		s: s,
	}

	req, err := middleware.RequestDoc(wirebson.MustDocument(
		"killCursors", "collection",
		"cursors", wirebson.RawArray([]byte{42}),
		"$db", "test",
	))
	require.NoError(t, err)

	resp, err := h.msgKillCursors(connCtx, req)
	require.Error(t, err)
	require.Nil(t, resp)
}
