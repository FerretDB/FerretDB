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

	"github.com/AlekSi/pointer"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCursor(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Scalars)

	opts := &options.FindOptions{
		BatchSize: pointer.ToInt32(1),
	}

	cur, err := collection.Find(ctx, bson.D{}, opts)
	require.NoError(t, err)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		res := []bson.D{}
		err := cur.All(ctx, &res)
		require.NoError(t, err)
	}()

	res := []bson.D{}
	err = cur.All(ctx, &res)
	require.NoError(t, err)

	wg.Wait()
}
