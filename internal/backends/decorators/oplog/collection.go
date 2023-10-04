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

package oplog

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/observability"
)

// collection implements backends.Collection interface by adding OpLog functionality to the wrapped collection.
type collection struct {
	c    backends.Collection
	db   *database
	name string
}

// newCollection creates a new collection that wraps the given collection.
func newCollection(c backends.Collection, name string, db *database) backends.Collection {
	return &collection{
		c:    c,
		db:   db,
		name: name,
	}
}

// Query implements backends.Collection interface.
func (c *collection) Query(ctx context.Context, params *backends.QueryParams) (*backends.QueryResult, error) {
	return c.c.Query(ctx, params)
}

// InsertAll implements backends.Collection interface.
func (c *collection) InsertAll(ctx context.Context, params *backends.InsertAllParams) (*backends.InsertAllResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := c.c.InsertAll(ctx, params)
	if err != nil {
		return nil, err
	}

	if c.name == "oplog.rs" {
		return res, nil
	}

	oplogDocs := make([]*types.Document, len(params.Docs))

	for i, doc := range params.Docs {
		oplogDoc, err := types.NewDocument(
			"_id", types.NewObjectID(), // TODO
			"ts", types.NextTimestamp(time.Now()),
			"ns", c.db.name+"."+c.name,
			"op", "i",
			"o", doc,
		)
		if err == nil {
			err = oplogDoc.ValidateData()
		}

		if err != nil {
			c.db.l.Error("Failed to create oplog document", zap.Error(err))
			return res, nil
		}

		oplogDocs[i] = oplogDoc
	}

	oplogC, err := c.db.Collection("oplog.rs")
	if err == nil {
		_, err = oplogC.InsertAll(ctx, &backends.InsertAllParams{
			Docs: oplogDocs,
		})
	}

	if err != nil {
		c.db.l.Error("Failed to insert oplog documents", zap.Error(err))
	}

	return res, nil
}

// UpdateAll implements backends.Collection interface.
func (c *collection) UpdateAll(ctx context.Context, params *backends.UpdateAllParams) (*backends.UpdateAllResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := c.c.UpdateAll(ctx, params)
	if err != nil {
		return nil, err
	}

	if c.name == "oplog.rs" {
		return res, nil
	}

	// TODO
	_ = res

	return res, nil
}

// DeleteAll implements backends.Collection interface.
func (c *collection) DeleteAll(ctx context.Context, params *backends.DeleteAllParams) (*backends.DeleteAllResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := c.c.DeleteAll(ctx, params)
	if err != nil {
		return nil, err
	}

	if c.name == "oplog.rs" {
		return res, nil
	}

	oplogDocs := make([]*types.Document, len(params.IDs))

	for i, id := range params.IDs {
		idDoc, err := types.NewDocument("_id", id)
		if err != nil {
			c.db.l.Error("Failed to create oplog _id document", zap.Error(err))
			return res, nil
		}

		oplogDoc, err := types.NewDocument(
			"_id", types.NewObjectID(), // TODO
			"ts", types.NextTimestamp(time.Now()),
			"ns", c.db.name+"."+c.name,
			"op", "d",
			"o", idDoc,
		)
		if err == nil {
			err = oplogDoc.ValidateData()
		}

		if err != nil {
			c.db.l.Error("Failed to create oplog document", zap.Error(err))
			return res, nil
		}

		oplogDocs[i] = oplogDoc
	}

	oplogC, err := c.db.Collection("oplog.rs")
	if err == nil {
		_, err = oplogC.InsertAll(ctx, &backends.InsertAllParams{
			Docs: oplogDocs,
		})
	}

	if err != nil {
		c.db.l.Error("Failed to insert oplog documents", zap.Error(err))
	}

	return res, nil
}

// Explain implements backends.Collection interface.
func (c *collection) Explain(ctx context.Context, params *backends.ExplainParams) (*backends.ExplainResult, error) {
	return c.c.Explain(ctx, params)
}

// Stats implements backends.Collection interface.
func (c *collection) Stats(ctx context.Context, params *backends.CollectionStatsParams) (*backends.CollectionStatsResult, error) {
	return c.c.Stats(ctx, params)
}

// Compact implements backends.Collection interface.
func (c *collection) Compact(ctx context.Context, params *backends.CompactParams) (*backends.CompactResult, error) {
	return c.c.Compact(ctx, params)
}

// ListIndexes implements backends.Collection interface.
func (c *collection) ListIndexes(ctx context.Context, params *backends.ListIndexesParams) (*backends.ListIndexesResult, error) {
	return c.c.ListIndexes(ctx, params)
}

// CreateIndexes implements backends.Collection interface.
func (c *collection) CreateIndexes(ctx context.Context, params *backends.CreateIndexesParams) (*backends.CreateIndexesResult, error) { //nolint:lll // for readability
	return c.c.CreateIndexes(ctx, params)
}

// DropIndexes implements backends.Collection interface.
func (c *collection) DropIndexes(ctx context.Context, params *backends.DropIndexesParams) (*backends.DropIndexesResult, error) {
	return c.c.DropIndexes(ctx, params)
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)
