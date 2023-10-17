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

package stages

import (
	"context"
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// collStats represents $collStats stage.
type collStats struct {
	storageStats   *storageStats
	count          bool
	latencyStats   bool
	queryExecStats bool
}

// storageStats represents $collStats.storageStats field.
type storageStats struct {
	scale int32
}

// newCollStats creates a new $collStats stage.
func newCollStats(stage *types.Document) (aggregations.Stage, error) {
	fields, err := common.GetRequiredParam[*types.Document](stage, "$collStats")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageCollStatsInvalidArg,
			fmt.Sprintf("$collStats must take a nested object but found: %s", types.FormatAnyValue(stage)),
			"$collStats (stage)",
		)
	}

	var cs collStats

	// Return error on invalid type of count.
	// TODO https://github.com/FerretDB/FerretDB/issues/2336
	cs.count = fields.Has("count")

	// TODO https://github.com/FerretDB/FerretDB/issues/2341
	cs.latencyStats = fields.Has("latencyStats")

	// TODO https://github.com/FerretDB/FerretDB/issues/2341
	cs.queryExecStats = fields.Has("queryExecStats")

	if fields.Has("storageStats") {
		cs.storageStats = new(storageStats)

		storageStatsFields := must.NotFail(fields.Get("storageStats")).(*types.Document)

		cs.storageStats.scale = 1

		var s any
		if s, err = storageStatsFields.Get("scale"); err == nil {
			scale, err := commonparams.GetValidatedNumberParamWithMinValue(
				"$collStats.storageStats", "scale", s, 1,
			)
			if err != nil {
				return nil, err
			}

			cs.storageStats.scale = int32(scale)
		}
	}

	return &cs, nil
}

// Process implements Stage interface.
//
// Processing consists of modification of the input document, so it contains all the necessary fields
// and the data is modified according to the given request.
func (c *collStats) Process(ctx context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error) { //nolint:lll // for readability
	_, res, err := iter.Next()
	if errors.Is(err, iterator.ErrIteratorDone) {
		// For non-shared collections, it must contain a single document.
		panic("collStatsStage: Process: expected 1 document, got none")
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if c.storageStats != nil {
		scale := c.storageStats.scale

		if c.storageStats.scale > 1 {
			scalable := []string{"size", "storageSize", "freeStorageSize", "totalIndexSize", "totalSize"}
			for _, key := range scalable {
				path := types.NewStaticPath("storageStats", key)
				val := must.NotFail(res.GetByPath(path))
				must.NoError(res.SetByPath(path, val.(int64)/int64(scale)))
			}

			path := types.NewStaticPath("storageStats", "indexSizes")
			indexSizesDoc := must.NotFail(res.GetByPath(path)).(*types.Document)

			for _, name := range indexSizesDoc.Keys() {
				size := must.NotFail(indexSizesDoc.Get(name)).(int64)
				indexSizesDoc.Set(name, size/int64(scale))
			}

			must.NoError(res.SetByPath(path, indexSizesDoc))
		}

		must.NoError(res.SetByPath(types.NewStaticPath("storageStats", "scaleFactor"), scale))
	}

	if _, _, err := iter.Next(); err == nil || !errors.Is(err, iterator.ErrIteratorDone) {
		// For non-shared collections, it contains only a single document.
		panic("collStatsStage: Process: expected 1 document, got more")
	}

	iter = iterator.Values(iterator.ForSlice([]*types.Document{res}))
	closer.Add(iter)

	return iter, nil
}

// check interfaces
var (
	_ aggregations.Stage = (*collStats)(nil)
)
