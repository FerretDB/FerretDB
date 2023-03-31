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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// collStats represents $collStats stage.
type collStats struct {
	fields *types.Document
}

// newCollStats creates a new $collStats stage.
func newCollStats(stage *types.Document) (Stage, error) {
	fields, err := common.GetRequiredParam[*types.Document](stage, "$collStats")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &collStats{
		fields: fields,
	}, nil
}

// Process implements Stage interface.
func (c *collStats) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	ns := "" // c.f.GetNameSpace()
	now := time.Now().UTC().Format(time.RFC3339)

	doc, err := types.NewDocument(
		"ns", ns,
		"host", host,
		"localTime", now,
	)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	count, err := c.fields.Get("count")

	if err == nil {
		doc.Set("count", int32(len(in)))
	}

	// TODO: return error on invalid type of count.
	// https://github.com/FerretDB/FerretDB/issues/2336
	_ = count

	// TODO: implement latencyStats
	// https://github.com/FerretDB/FerretDB/issues/1416
	latencyStats, _ := c.fields.Get("latencyStats")
	_ = latencyStats

	// TODO: implement queryExecStats
	// https://github.com/FerretDB/FerretDB/issues/1416
	queryExecStats, _ := c.fields.Get("queryExecStats")
	_ = queryExecStats

	// TODO: implement storageStats
	// https://github.com/FerretDB/FerretDB/issues/1416
	storageStats, _ := c.fields.Get("storageStats")
	_ = storageStats

	return []*types.Document{doc}, nil
}

// check interfaces
var (
	_ Stage = (*collStats)(nil)
)
