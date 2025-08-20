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
	"sync"
	"testing"
	"time"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestSleepParallelWrite(t *testing.T) {
	ctx, coll := setup.Setup(t)
	adminDB := coll.Database().Client().Database("admin")

	var wg sync.WaitGroup

	start := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()

		var res bson.D

		timeBefore := time.Now()
		close(start)
		err := adminDB.RunCommand(ctx, bson.D{
			{"sleep", int32(1)},
			{"millis", int32(2000)},
		}).Decode(&res)

		dur := time.Since(timeBefore)
		AssertEqualDocuments(t, bson.D{{"ok", float64(1.0)}}, res)

		assert.InDelta(t, 2000, dur.Milliseconds(), 100)
		require.NoError(t, err)
	}()

	<-start
	timeBefore := time.Now()
	_, err := coll.InsertOne(ctx, bson.D{{"foo", 1}})

	dur := time.Since(timeBefore)

	assert.InDelta(t, 2000, dur.Milliseconds(), 100)
	require.NoError(t, err)

	wg.Wait()
}

func TestSleepParallelRead(tt *testing.T) {
	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/5399")

	ctx, coll := setup.Setup(t)
	adminDB := coll.Database().Client().Database("admin")

	var wg sync.WaitGroup

	start := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()

		var res bson.D

		timeBefore := time.Now()
		close(start)
		err := adminDB.RunCommand(ctx, bson.D{
			{"sleep", int32(1)},
			{"millis", int32(2000)},
		}).Decode(&res)

		dur := time.Since(timeBefore)
		AssertEqualDocuments(t, bson.D{{"ok", float64(1.0)}}, res)

		assert.InDelta(t, 2000, dur.Milliseconds(), 100)
		require.NoError(t, err)
	}()

	<-start
	timeBefore := time.Now()

	_, err := coll.Find(ctx, bson.D{})
	require.NoError(t, err)

	dur := time.Since(timeBefore)

	assert.InDelta(t, 0, dur.Milliseconds(), 100)
	require.NoError(t, err)

	wg.Wait()
}
