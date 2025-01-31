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
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/FerretDB/wire/wirebson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestCurrentOpGetMore(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, new(setup.SetupOpts))
	ctx, collection, cName, dbName := s.Ctx, s.Collection, s.Collection.Name(), s.Collection.Database().Name()
	adminDB := collection.Database().Client().Database("admin")

	var docs []any
	for x := range 200 {
		docs = append(docs, bson.D{
			{"_id", int64(x)},
			{"user", fmt.Sprintf("user%d", x)},
			{"even", x%2 == 0},
		})
	}

	_, err := collection.InsertMany(ctx, docs)
	require.NoError(t, err)

	searchF := func(op *wirebson.Document) bool {
		// find an in-progress operation started by this test
		return op.Get("op") == "getmore" && op.Get("ns") == cName+"."+dbName
	}

	res := inProgress(t, ctx, adminDB, 5, searchF, func(ready chan<- struct{}, start <-chan struct{}) {
		cursor, err := collection.Find(ctx, bson.D{}) //nolint:vet // redeclare err to avoid datarace
		require.NoError(t, err)

		cursor.SetBatchSize(1)

		ready <- struct{}{}

		<-start

		var res []bson.D
		assert.NoError(t, cursor.All(ctx, &res))
		assert.NotEmpty(t, res)
	})

	inprog, err := res.Get("inprog").(wirebson.AnyArray).Decode()
	require.NoError(t, err)

	var getMoreOp *wirebson.Document

	for v := range inprog.Values() {
		getMoreOp, err = v.(wirebson.AnyDocument).Decode()
		require.NoError(t, err)

		if searchF(getMoreOp) {
			break
		}
	}

	require.NotNil(t, getMoreOp, "no in-progress getMore operation found")

	currentOpTime, ok := getMoreOp.Get("currentOpTime").(string)
	require.True(t, ok, wirebson.LogMessageIndent(getMoreOp))
	_, err = time.Parse(time.RFC3339, currentOpTime)
	require.NoError(t, err)

	opid, ok := getMoreOp.Get("opid").(int32)
	require.True(t, ok)
	require.NotZero(t, opid, wirebson.LogMessageIndent(getMoreOp))

	v, ok := getMoreOp.Get("command").(wirebson.AnyDocument)
	require.True(t, ok)
	command, err := v.Decode()
	require.NoError(t, err)

	cursorID := command.Get("getMore")
	require.NotNil(t, cursorID, wirebson.LogMessageIndent(getMoreOp))

	lsid := command.Get("lsid")
	require.NotNil(t, lsid, wirebson.LogMessageIndent(getMoreOp))

	// command contains MongoDB specific fields
	fixCluster(t, command)
	err = getMoreOp.Replace("command", command)
	require.NoError(t, err)

	// optional fields
	getMoreOp.Remove("secs_running")
	getMoreOp.Remove("microsecs_running")

	// unsupported fields
	getMoreOp.Remove("host")
	getMoreOp.Remove("desc")
	getMoreOp.Remove("connectionId")
	getMoreOp.Remove("client")
	getMoreOp.Remove("clientMetadata")
	getMoreOp.Remove("effectiveUsers")
	getMoreOp.Remove("threaded")
	getMoreOp.Remove("lsid")
	getMoreOp.Remove("numYields")
	getMoreOp.Remove("redacted")
	getMoreOp.Remove("locks")
	getMoreOp.Remove("waitingForLock")
	getMoreOp.Remove("lockStats")
	getMoreOp.Remove("waitingForFlowControl")
	getMoreOp.Remove("flowControlStats")
	getMoreOp.Remove("queryFramework")
	getMoreOp.Remove("planSummary")
	getMoreOp.Remove("cursor")

	// MongoDB may contain other in-progress operations
	err = res.Replace("inprog", wirebson.MustArray(getMoreOp))
	require.NoError(t, err)

	fixCluster(t, res)

	expected := wirebson.MustDocument(
		"inprog", wirebson.MustArray(wirebson.MustDocument(
			"type", "op",
			"active", true,
			"currentOpTime", currentOpTime,
			"opid", opid,
			"op", "getmore",
			"ns", cName+"."+dbName,
			"command", wirebson.MustDocument(
				"getMore", cursorID,
				"collection", cName,
				"batchSize", int32(1),
				"lsid", lsid,
				"$db", dbName,
			))),
		"ok", 1.0,
	)

	testutil.AssertEqual(t, expected, res)
}

func TestCurrentOpExplain(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, new(setup.SetupOpts))
	ctx, collection, cName, dbName := s.Ctx, s.Collection, s.Collection.Name(), s.Collection.Database().Name()
	adminDB := collection.Database().Client().Database("admin")

	var docs []any
	for x := range 200 {
		docs = append(docs, bson.D{
			{"_id", int64(x)},
			{"user", fmt.Sprintf("user%d", x)},
			{"even", x%2 == 0},
		})
	}

	_, err := collection.InsertMany(ctx, docs)
	require.NoError(t, err)

	searchF := func(op *wirebson.Document) bool {
		// find an in-progress operation started by this test
		return op.Get("op") == "command" && op.Get("ns") == cName+"."+dbName
	}

	res := inProgress(t, ctx, adminDB, 500, searchF, func(ready chan<- struct{}, start <-chan struct{}) {
		command := bson.D{
			{"explain", bson.D{
				{"find", collection.Name()},
			}},
		}

		ready <- struct{}{}

		<-start

		err := collection.Database().RunCommand(ctx, command).Err() //nolint:vet // redeclare err to avoid datarace
		require.NoError(t, err)
	})

	inprog, err := res.Get("inprog").(wirebson.AnyArray).Decode()
	require.NoError(t, err)

	var op *wirebson.Document

	for v := range inprog.Values() {
		op, err = v.(wirebson.AnyDocument).Decode()
		require.NoError(t, err)

		if searchF(op) {
			break
		}
	}

	require.NotNil(t, op, "no in-progress explain operation found")

	currentOpTime, ok := op.Get("currentOpTime").(string)
	require.True(t, ok, wirebson.LogMessageIndent(op))
	_, err = time.Parse(time.RFC3339, currentOpTime)
	require.NoError(t, err)

	opid, ok := op.Get("opid").(int32)
	require.True(t, ok)
	require.NotZero(t, opid, wirebson.LogMessageIndent(op))

	v, ok := op.Get("command").(wirebson.AnyDocument)
	require.True(t, ok)
	command, err := v.Decode()
	require.NoError(t, err)

	lsid := command.Get("lsid")
	require.NotNil(t, lsid, wirebson.LogMessageIndent(op))

	// command contains MongoDB specific fields
	fixCluster(t, command)
	err = op.Replace("command", command)
	require.NoError(t, err)

	// optional fields
	op.Remove("secs_running")
	op.Remove("microsecs_running")

	// unsupported fields
	op.Remove("host")
	op.Remove("desc")
	op.Remove("connectionId")
	op.Remove("client")
	op.Remove("clientMetadata")
	op.Remove("effectiveUsers")
	op.Remove("threaded")
	op.Remove("lsid")
	op.Remove("numYields")
	op.Remove("redacted")
	op.Remove("locks")
	op.Remove("waitingForLock")
	op.Remove("lockStats")
	op.Remove("waitingForFlowControl")
	op.Remove("flowControlStats")
	op.Remove("queryFramework")
	op.Remove("planSummary")
	op.Remove("cursor")

	// MongoDB may contain other in-progress operations
	err = res.Replace("inprog", wirebson.MustArray(op))
	require.NoError(t, err)

	fixCluster(t, res)

	expected := wirebson.MustDocument(
		"inprog", wirebson.MustArray(wirebson.MustDocument(
			"type", "op",
			"active", true,
			"currentOpTime", currentOpTime,
			"opid", opid,
			"op", "command",
			"ns", cName+"."+dbName,
			"command", wirebson.MustDocument(
				"explain", wirebson.MustDocument(
					"find", cName),
				"lsid", lsid,
				"$db", dbName,
			))),
		"ok", 1.0,
	)

	testutil.AssertEqual(t, expected, res)
}

// inProgress runs function `f` in `n` go routines while `currentOp` is executed to search an in-progress operation.
// It repeats until it finds an in-progress operation that matches the search function
// or reaches the attempts limit.
// It returns the response of the `currentOp` command.
func inProgress(tb testing.TB, ctx context.Context, adminDB *mongo.Database, n int, searchF func(op *wirebson.Document) bool, f func(ready chan<- struct{}, start <-chan struct{})) *wirebson.Document { //nolint:lll // for readability
	// attempts is chosen to prevent the test failing while not taking too long
	attempts := 10

	// nCurrentOp is chosen to run currentOp command every millisecond apart to capture in-progress operation
	nCurrentOp := 250

	for range attempts {
		var wg sync.WaitGroup
		readyCh := make(chan struct{}, n+nCurrentOp)
		startCh := make(chan struct{})

		for i := 0; i < n; i++ {
			wg.Add(1)

			go func() {
				defer func() {
					wg.Done()
				}()

				f(readyCh, startCh)
			}()
		}

		resCh := make(chan *wirebson.Document, nCurrentOp)

		// repeat `currentOp` command to find an in-progress operation that matches the `search` function
		for i := 0; i < nCurrentOp; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				readyCh <- struct{}{}

				<-startCh

				time.Sleep(time.Duration(i) * time.Millisecond)

				var actual bson.D
				err := adminDB.RunCommand(ctx, bson.D{{"currentOp", 1}}).Decode(&actual) //nolint:vet // redeclare err to avoid datarace
				require.NoError(tb, err)

				doc, err := convert(tb, actual).(wirebson.AnyDocument).Decode()
				require.NoError(tb, err)

				v := doc.Get("inprog")
				require.NotNil(tb, v)

				inprog, err := v.(wirebson.AnyArray).Decode()
				require.NoError(tb, err)

				for v = range inprog.Values() {
					var op *wirebson.Document

					op, err = v.(wirebson.AnyDocument).Decode()
					require.NoError(tb, err)

					if searchF(op) {
						resCh <- doc
						return
					}
				}
			}()
		}

		for i := 0; i < len(readyCh); i++ {
			if _, ok := <-readyCh; !ok {
				tb.Error("closed ready channel")
			}
		}

		close(startCh)

		wg.Wait()

		close(resCh)

		if res, ok := <-resCh; ok {
			return res
		}
	}

	require.Fail(tb, "no in-progress operation found")

	panic("unreachable")
}
