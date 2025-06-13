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

package query

import (
	"testing"
	"time"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestInsertZeroTimestamp(tt *testing.T) {
	tt.Parallel()

	var t testing.TB = tt

	s := setup.SetupWithOpts(t, &setup.SetupOpts{WireConn: setup.WireConnAuth})

	_, resp, err := s.WireConn.Request(s.Ctx, wire.MustOpMsg(
		"insert", s.Collection.Name(),
		"documents", wirebson.MustArray(
			wirebson.MustDocument(
				"_id", "id",
				"v", wirebson.Timestamp(0),
				"d", wirebson.MustDocument(
					"dv", wirebson.Timestamp(0),
				),
			),
		),
		"$db", s.Collection.Database().Name(),
	))
	require.NoError(t, err)

	actual, err := resp.(*wire.OpMsg).Document()
	require.NoError(t, err)
	require.Equal(t, 1.0, actual.Get("ok"))

	_, resp, err = s.WireConn.Request(s.Ctx, wire.MustOpMsg(
		"find", s.Collection.Name(),
		"$db", s.Collection.Database().Name(),
	))
	require.NoError(t, err)

	actual, err = resp.(*wire.OpMsg).DocumentDeep()
	require.NoError(t, err)
	require.Equal(t, 1.0, actual.Get("ok"))

	batch := actual.Get("cursor").(*wirebson.Document).Get("firstBatch").(*wirebson.Array)

	d := batch.Get(0).(*wirebson.Document).Get("d").(*wirebson.Document)
	assert.Equal(t, wirebson.Timestamp(0), d.Get("dv"))

	// TODO https://github.com/FerretDB/FerretDB/issues/1608
	t = setup.FailsForFerretDB(t, "https://github.com/FerretDB/FerretDB/issues/1608")

	v := batch.Get(0).(*wirebson.Document).Get("v").(wirebson.Timestamp)
	assert.NotEqual(t, wirebson.Timestamp(0), v)
	assert.NotZero(t, v.I())
	assert.InDelta(t, time.Now().Unix(), v.T(), 5.0)
}

func TestInsertZeroTimestampBypass(tt *testing.T) {
	tt.Parallel()

	var t testing.TB = tt

	s := setup.SetupWithOpts(t, &setup.SetupOpts{WireConn: setup.WireConnAuth})

	_, resp, err := s.WireConn.Request(s.Ctx, wire.MustOpMsg(
		"insert", s.Collection.Name(),
		"documents", wirebson.MustArray(
			wirebson.MustDocument(
				"_id", "id",
				"v", wirebson.Timestamp(0),
				"d", wirebson.MustDocument(
					"dv", wirebson.Timestamp(0),
				),
			),
		),
		"bypassEmptyTsReplacement", true,
		"$db", s.Collection.Database().Name(),
	))
	require.NoError(t, err)

	actual, err := resp.(*wire.OpMsg).Document()
	require.NoError(t, err)
	require.Equal(t, 1.0, actual.Get("ok"))

	_, resp, err = s.WireConn.Request(s.Ctx, wire.MustOpMsg(
		"find", s.Collection.Name(),
		"$db", s.Collection.Database().Name(),
	))
	require.NoError(t, err)

	actual, err = resp.(*wire.OpMsg).DocumentDeep()
	require.NoError(t, err)
	require.Equal(t, 1.0, actual.Get("ok"))

	batch := actual.Get("cursor").(*wirebson.Document).Get("firstBatch").(*wirebson.Array)

	d := batch.Get(0).(*wirebson.Document).Get("d").(*wirebson.Document)
	assert.Equal(t, wirebson.Timestamp(0), d.Get("dv"))

	v := batch.Get(0).(*wirebson.Document).Get("v").(wirebson.Timestamp)
	assert.Equal(t, wirebson.Timestamp(0), v)
}
