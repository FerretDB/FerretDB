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

package aggregations

import (
	"context"
	"os"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// CreatePipelineParams are parameters use for creating pipeline.
type CreatePipelineParams struct {
	DB         string
	Collection string
}

// CreatePipeline creates a pipeline by querying necessary data from the database if needed.
// The first stage (or first few stages) of the pipeline impacts how initial iterator is created for the pipeline.
//
//nolint:lll // demonstration purpose only
func CreatePipeline(ctx context.Context, params CreatePipelineParams, aggregation Aggregation, closer *iterator.MultiCloser, stagesDocs []any) (types.DocumentsIterator, error) {
	var initialStage, secondStage *types.Document

	if len(stagesDocs) > 0 {
		initialStage = stagesDocs[0].(*types.Document)
	}

	if len(stagesDocs) > 1 {
		secondStage = stagesDocs[1].(*types.Document)
	}

	_ = secondStage

	var match, sort *types.Document

	switch {
	case initialStage == nil:
		return nil, nil
	case initialStage.Has("$collStats"):
		fields := must.NotFail(initialStage.Get("$collStats")).(*types.Document)
		return collStats(ctx, params, aggregation, closer, fields)
		// case initialStage.Has("$match"):
		//	// this is potentially pushdown query can be set
		//	// only if we want to control query pushdown at handler level
		//	match = must.NotFail(initialStage.Get("$match")).(*types.Document)
		//	if secondStage.Has("$sort") {
		//		sort = must.NotFail(initialStage.Get("$sort")).(*types.Document)
		//	}
		// case initialStage.Has("$sort"):
		//	// this is potentially pushdown query can be set
		//	// only if we want to control query pushdown at handler level
		//	sort = must.NotFail(initialStage.Get("$sort")).(*types.Document)
		//	if secondStage.Has("$sort") {
		//		match = must.NotFail(initialStage.Get("$match")).(*types.Document)
		//	}
	}

	return aggregation.Query(ctx, QueryParams{
		Filter: match,
		Sort:   sort,
	}, closer)
}

//nolint:lll // demonstration purpose only
func collStats(ctx context.Context, params CreatePipelineParams, aggregation Aggregation, closer *iterator.MultiCloser, fields *types.Document) (types.DocumentsIterator, error) {
	var host string
	var err error

	host, err = os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc := must.NotFail(types.NewDocument(
		"ns", params.DB+"."+params.Collection,
		"host", host,
		"localTime", time.Now().UTC().Format(time.RFC3339),
	))

	var stats *CollStatsResult

	if fields.Has("count") || fields.Has("storageStats") {
		stats, err = aggregation.CollStats(ctx, closer)
		if err != nil {
			return nil, err
		}
	}

	if fields.Has("storageStats") {
		var avgObjSize int64
		if stats.CountObjects > 0 {
			avgObjSize = stats.SizeCollection / stats.CountObjects
		}

		doc.Set(
			"storageStats", must.NotFail(types.NewDocument(
				"size", stats.SizeTotal,
				"count", stats.CountObjects,
				"avgObjSize", avgObjSize,
				"storageSize", stats.SizeCollection,
				"freeStorageSize", int64(0), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"capped", false, // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"wiredTiger", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"nindexes", stats.CountIndexes,
				"indexDetails", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"indexBuilds", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
				"totalIndexSize", stats.SizeIndexes,
				"totalSize", stats.SizeTotal,
				"indexSizes", must.NotFail(types.NewDocument()), // TODO https://github.com/FerretDB/FerretDB/issues/2342
			)),
		)
	}

	if fields.Has("count") {
		doc.Set(
			"count", stats.CountObjects,
		)
	}

	// Process the retrieved statistics through the stages.
	iter := iterator.Values(iterator.ForSlice([]*types.Document{doc}))
	closer.Add(iter)

	return iter, nil
}
