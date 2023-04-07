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
	"errors"
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

	for {
		docs, err := iterator.ConsumeValuesN(iter, 10)
		if errors.Is(err, iterator.ErrIteratorDone) {
			// TODO insert leftovers
			break
		}
		require.NoError(tb, err)

		var insertDocs []any = make([]any, len(docs))
		for i, d := range docs {
			insertDocs[i] = d
		}

		_, err = coll.InsertMany(ctx, insertDocs)
		require.NoError(tb, err)
	}

	return ctx, coll
}
