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
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestRenameCollectionStress(t *testing.T) {
	// TODO rewrite using teststress.Stress

	setup.SkipForTigrisWithReason(t, "Command renameCollection is not supported for Tigris")

	ctx, collection := setup.Setup(t) // no providers there, we will create collections for this test
	db := collection.Database()

	// Collection rename should be performed while connecting to the admin database.
	adminDB := db.Client().Database("admin")

	collNum := runtime.GOMAXPROCS(-1) * 10

	// create collections that we will attempt to rename later
	for i := 0; i < collNum; i++ {
		collName := fmt.Sprintf("rename_collection_stress_%d", i)
		err := db.CreateCollection(ctx, collName)
		require.NoError(t, err)
	}

	ready := make(chan struct{}, collNum)
	start := make(chan struct{})

	var errNum atomic.Int32
	var renamedCollectionsNum atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < collNum; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			ready <- struct{}{}

			<-start

			renameFrom := fmt.Sprintf("%s.rename_collection_stress_%d", db.Name(), i)
			renameTo := fmt.Sprintf("%s.rename_collection_stress_renamed", db.Name())

			var res bson.D
			err := adminDB.RunCommand(ctx,
				bson.D{
					{"renameCollection", renameFrom},
					{"to", renameTo},
				},
			).Decode(&res)

			if err != nil {
				errNum.Add(1)
			} else {
				renamedCollectionsNum.Add(1)
			}
		}(i)
	}

	for i := 0; i < collNum; i++ {
		<-ready
	}

	close(start)

	wg.Wait()

	val := atomic.Int32{}
	val.Add(int32(collNum - 1))
	require.Equal(t, val.Load(), errNum.Load())              // expected to have errors for all the rename attempts apart from the first one
	require.Equal(t, int32(1), renamedCollectionsNum.Load()) // expected to have one collection renamed

	colls, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	require.Contains(t, colls, "rename_collection_stress_renamed")
}
