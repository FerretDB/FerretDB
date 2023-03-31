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

package pg

import (
	"context"
	"os"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// collStats represents $collStats stage.
type collStats struct {
	fields *types.Document
	dbPool *pgdb.Pool
	qp     pgdb.QueryParams
}

// newCollStats creates a new $collStats stage.
func newCollStats(stage *types.Document, dbPool *pgdb.Pool, qp pgdb.QueryParams) (aggregations.Stage, error) {
	fields, err := common.GetRequiredParam[*types.Document](stage, "$collStats")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &collStats{
		dbPool: dbPool,
		qp:     qp,
		fields: fields,
	}, nil
}

// Process implements Stage interface.
func (c *collStats) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := types.NewDocument(
		"ns", c.qp.DB+"."+c.qp.Collection,
		"host", host,
		"localtime", time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO: check for the count value type and return error on invalid type
	if c.fields.Has("count") {
		doc.Set("count", int32(len(in)))
	}

	return []*types.Document{doc}, nil
}

// check interfaces
var (
	_ aggregations.Stage = (*collStats)(nil)
)
