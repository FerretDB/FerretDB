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
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

func TestRenameCollectionStress(t *testing.T) {
	ctx, collection := setup.Setup(t)
	db := collection.Database()

	adminDB := db.Client().Database("admin")

	var i, oks, errs atomic.Int32

	teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
		from := "rename_collection_stress_" + strconv.Itoa(int(i.Add(1)))
		err := db.CreateCollection(ctx, from)
		require.NoError(t, err)

		from = db.Name() + "." + from
		to := db.Name() + "." + "rename_collection_stress_renamed"

		ready <- struct{}{}
		<-start

		err = adminDB.RunCommand(ctx, bson.D{{"renameCollection", from}, {"to", to}}).Err()
		if err == nil {
			t.Logf("Renamed %q to %q.", from, to)
			oks.Add(1)
		} else {
			t.Log(err)
			errs.Add(1)
		}
	})

	assert.Equal(t, int32(1), oks.Load())
	assert.Equal(t, i.Load()-1, errs.Load())

	colls, err := db.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)
	require.Contains(t, colls, "rename_collection_stress_renamed")
}
