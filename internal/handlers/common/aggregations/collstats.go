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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
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
func newCollStats(stage *types.Document) (Stage, error) {
	fields, err := common.GetRequiredParam[*types.Document](stage, "$collStats")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageCollStatsInvalidArg,
			fmt.Sprintf("$collStats must take a nested object but found: $collStats: %s", types.FormatAnyValue(stage)),
			"$collStats (stage)",
		)
	}

	var cs collStats

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
func (c *collStats) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	// The result of $collStats stage is always an array with a single document, and we expect the same input.
	if len(in) != 1 {
		panic(fmt.Sprintf("collStats: expected 1 document, got %d", len(in)))
	}

	res := in[0]

	if c.storageStats != nil && c.storageStats.scale > 1 {
		scale := float64(c.storageStats.scale)

		scalable := []string{"size", "storageSize", "freeStorageSize", "totalIndexSize"}
		for _, key := range scalable {
			if res.Has(key) {
				res.Set(key, must.NotFail(res.Get(key)).(float64)/scale)
			}
		}
	}

	return []*types.Document{res}, nil
}

// Type implements Stage interface.
func (c *collStats) Type() StageType {
	return StageTypeStats
}

// check interfaces
var (
	_ Stage = (*collStats)(nil)
)
