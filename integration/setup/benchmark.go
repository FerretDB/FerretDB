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
	"testing"

	"github.com/FerretDB/FerretDB/integration/benchmarkdata"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type SetupBenchmarkResults struct {
	Ctx                        context.Context
	TargetCollection           *mongo.Collection
	TargetNoPushdownCollection *mongo.Collection
	CompatCollection           *mongo.Collection
}

// SetupBenchmark
func SetupBenchmark(tb testing.TB, insertData benchmarkdata.Data) *SetupBenchmarkResults {
	tb.Helper()

	if *compatURLF == "" {
		tb.Skip("-compat-url is empty, skipping benchmark")
	}

	ctx, cancel := context.WithCancel(testutil.Ctx(tb))

	level := zap.NewAtomicLevelAt(zap.ErrorLevel)
	if *debugSetupF {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger := testutil.Logger(tb, level)

	targetClient, _ := setupListener(tb, ctx, logger)
	targetNoPushdownClient, _ := setupListenerWithOpts(tb, ctx, &setupListenerOpts{
		Logger:          logger,
		DisablePushdown: true,
	})

	compatClient := setupClient(tb, ctx, *compatURLF)

	tb.Cleanup(cancel)

	var collections []*mongo.Collection

	for _, client := range []*mongo.Client{
		targetClient,
		targetNoPushdownClient,
		compatClient,
	} {
		coll := setupCollection(tb, ctx, client, &SetupOpts{})
		require.NoError(tb, insertData(ctx, coll))
		collections = append(collections, coll)
	}

	return &SetupBenchmarkResults{
		Ctx:                        ctx,
		TargetCollection:           collections[0],
		TargetNoPushdownCollection: collections[1],
		CompatCollection:           collections[2],
	}
}
