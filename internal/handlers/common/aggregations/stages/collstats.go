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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// collStatsStage represents $collStats stage.
type collStatsStage struct {
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
func newCollStats(stage *types.Document) (Stage, error) {
	fields, err := common.GetRequiredParam[*types.Document](stage, "$collStats")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageCollStatsInvalidArg,
			fmt.Sprintf("$collStats must take a nested object but found: %s", types.FormatAnyValue(stage)),
			"$collStats (stage)",
		)
	}

	var cs collStatsStage

	// TODO Return error on invalid type of count: https://github.com/FerretDB/FerretDB/issues/2336
	cs.count = fields.Has("count")

	// TODO Implement latencyStats: https://github.com/FerretDB/FerretDB/issues/2341
	cs.latencyStats = fields.Has("latencyStats")

	// TODO Implement queryExecStats: https://github.com/FerretDB/FerretDB/issues/2341
	cs.queryExecStats = fields.Has("queryExecStats")

	if fields.Has("storageStats") {
		cs.storageStats = new(storageStats)

		// TODO Add proper support for scale: https://github.com/FerretDB/FerretDB/issues/1346
		cs.storageStats.scale, err = common.GetOptionalPositiveNumber(
			must.NotFail(fields.Get("storageStats")).(*types.Document),
			"scale",
		)
		if err != nil || cs.storageStats.scale == 0 {
			cs.storageStats.scale = 1
		}
	}

	return &cs, nil
}

// Process implements Stage interface.
//
// Processing consists of modification of the input document, so it contains all the necessary fields
// and the data is modified according to the given request.
func (c *collStatsStage) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	// For non-shared collections, the input must be an array with a single document.
	if len(in) != 1 {
		panic(fmt.Sprintf("collStatsStage: Process: expected 1 document, got %d", len(in)))
	}

	res := in[0]

	if c.storageStats != nil {
		scale := c.storageStats.scale

		if c.storageStats.scale > 1 {
			scalable := []string{"size", "avgObjSize", "storageSize", "freeStorageSize", "totalIndexSize"}
			for _, key := range scalable {
				path := types.NewStaticPath("storageStats", key)
				val := must.NotFail(res.GetByPath(path))
				must.NoError(res.SetByPath(path, val.(int32)/scale))
			}
		}

		must.NoError(res.SetByPath(types.NewStaticPath("storageStats", "scaleFactor"), scale))
	}

	return []*types.Document{res}, nil
}

// Type implements Stage interface.
func (c *collStatsStage) Type() StageType {
	return StageTypeStats
}

// check interfaces
var (
	_ Stage = (*collStatsStage)(nil)
)
