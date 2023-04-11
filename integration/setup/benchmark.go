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

// Package setup provides integration tests setup helpers.
package setup

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupBenchmark sets up in-process FerretDB targets with pushdown enabled and disabled,
// establishes a connection to the compat database, and returns collections for each database.
func SetupBenchmark(tb testing.TB, p shareddata.BenchmarkProvider) (context.Context, *mongo.Collection) {
	tb.Helper()

	s := SetupWithOpts(tb, &SetupOpts{})

	ctx := s.Ctx
	coll := s.Collection

	iter := p.Docs()
	defer iter.Close()

	var done bool

	hash := sha256.New()

	for !done {
		docs, err := iterator.ConsumeValuesN(iter, 10)
		if errors.Is(err, iterator.ErrIteratorDone) {
			done = true
		}
		require.NoError(tb, err)

		var insertDocs []any = make([]any, len(docs))
		for i, doc := range docs {

			rawDoc := []byte(fmt.Sprintf("%x", doc))
			currSum := hash.Sum(rawDoc)
			hash.Reset()

			if _, err := hash.Write(currSum); err != nil {
				panic("Unexpected error: " + err.Error())
			}

			insertDocs[i] = doc
		}

		if len(docs) == 0 {
			break
		}

		_, err = coll.InsertMany(ctx, insertDocs)
		require.NoError(tb, err)
	}

	actualSum := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	require.Equal(tb, p.Hash(), actualSum, "The checksum of inserted documents in BenchmarkProvider is different than specified.",
		"If you didn't change any data, it could mean that provider generates different data or provides it in different order.")

	return ctx, coll
}
