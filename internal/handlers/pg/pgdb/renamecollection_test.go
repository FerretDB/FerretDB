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

package pgdb

/*

func TestRenameCollectionStress(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Rename is not supported for Tigris")
	// no providers needed as we will create collections ourselves.
	// but CI complains about no providers so we will use shareddata.Bools.
	s := setup.SetupWithOpts(
		t, &setup.SetupOpts{
			DatabaseName: "admin", Providers: []shareddata.Provider{shareddata.Bools},
		},
	)

	ctx, collection := s.Ctx, s.Collection

	db := collection.Database()

	collNum := runtime.GOMAXPROCS(-1) * 10

	// create collections that we will attempt to rename later
	for i := 0; i < collNum; i++ {
		collName := fmt.Sprintf("rename_collection_stress_%d", i)
		err := db.CreateCollection(ctx, collName)
		require.NoError(t, err)
	}

	ready := make(chan struct{}, collNum)
	start := make(chan struct{})

	// we can't use a sync.Map because it doesn't have a Len method.
	var errNum atomic.Int32
	renamedCollections := map[string]struct{}{}
	mx := new(sync.Mutex)

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
			err := db.RunCommand(ctx,
				bson.D{
					{"renameCollection", renameFrom},
					{"to", renameTo},
				},
			).Decode(&res)

			if err != nil {
				errNum.Add(1)
			} else {
				mx.Lock()
				renamedCollections[renameFrom] = struct{}{}
				mx.Unlock()
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
	require.Equal(t, val.Load(), errNum.Load()) // expected to have errors for all the rename attempts apart from the first one

	mx.Lock()
	renamedCollectionsCount := len(renamedCollections)
	mx.Unlock()

	require.Equal(t, 1, renamedCollectionsCount)

	colls, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	var found bool

	for _, coll := range colls {
		if coll == "rename_collection_stress_renamed" {
			found = true
			break
		}
	}

	require.True(t, found)

	// dropping the admin database is prohibited, so we do this instead.
	// unsure if necessary for CI, but it's a good idea to clean up after ourselves.
	for _, coll := range colls {
		if coll == "system.version" {
			continue
		}

		require.NoError(t, db.Collection(coll).Drop(ctx))
	}
}
*/
