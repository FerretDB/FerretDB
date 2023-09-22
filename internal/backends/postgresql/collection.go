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

package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// collection implements backends.Collection interface.
type collection struct {
	r      *metadata.Registry
	dbName string
	name   string
}

// newCollection creates a new Collection.
func newCollection(r *metadata.Registry, dbName, name string) backends.Collection {
	return backends.CollectionContract(&collection{
		r:      r,
		dbName: dbName,
		name:   name,
	})
}

// Query implements backends.Collection interface.
func (c *collection) Query(ctx context.Context, params *backends.QueryParams) (*backends.QueryResult, error) {
	p, err := c.r.DatabaseGetExisting(ctx, c.dbName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if p == nil {
		return &backends.QueryResult{
			Iter: newQueryIterator(ctx, nil),
		}, nil
	}

	meta, err := c.r.CollectionGet(ctx, c.dbName, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if meta == nil {
		return &backends.QueryResult{
			Iter: newQueryIterator(ctx, nil),
		}, nil
	}

	var placeholder Placeholder
	var whereClause string
	var args []any

	if params.Filter.Len() == 1 {
		if whereClause, args, err = prepareWhereClause(&placeholder, params.Filter); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var orderByClause string

	if params.Sort != nil {
		var sortArgs []any

		orderByClause, sortArgs, err = prepareOrderByClause(&placeholder, params.Sort)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		args = append(args, sortArgs...)
	}

	q := fmt.Sprintf(
		`SELECT %s FROM %s %s %s`,
		metadata.DefaultColumn,
		pgx.Identifier{c.dbName, meta.TableName}.Sanitize(),
		whereClause,
		orderByClause,
	)

	rows, err := p.Query(ctx, q, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &backends.QueryResult{
		Iter:           newQueryIterator(ctx, rows),
		FilterPushdown: whereClause != "",
		SortPushdown:   orderByClause != "",
	}, nil
}

// InsertAll implements backends.Collection interface.
func (c *collection) InsertAll(ctx context.Context, params *backends.InsertAllParams) (*backends.InsertAllResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3395
	return new(backends.InsertAllResult), nil
}

// UpdateAll implements backends.Collection interface.
func (c *collection) UpdateAll(ctx context.Context, params *backends.UpdateAllParams) (*backends.UpdateAllResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3391
	return new(backends.UpdateAllResult), nil
}

// DeleteAll implements backends.Collection interface.
func (c *collection) DeleteAll(ctx context.Context, params *backends.DeleteAllParams) (*backends.DeleteAllResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3400
	return new(backends.DeleteAllResult), nil
}

// Explain implements backends.Collection interface.
func (c *collection) Explain(ctx context.Context, params *backends.ExplainParams) (*backends.ExplainResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3389
	return new(backends.ExplainResult), nil
}

// Stats implements backends.Collection interface.
func (c *collection) Stats(ctx context.Context, params *backends.CollectionStatsParams) (*backends.CollectionStatsResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3398
	return new(backends.CollectionStatsResult), nil
}

// ListIndexes implements backends.Collection interface.
func (c *collection) ListIndexes(ctx context.Context, params *backends.ListIndexesParams) (*backends.ListIndexesResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3394
	return new(backends.ListIndexesResult), nil
}

// CreateIndexes implements backends.Collection interface.
func (c *collection) CreateIndexes(ctx context.Context, params *backends.CreateIndexesParams) (*backends.CreateIndexesResult, error) { //nolint:lll // for readability
	// TODO https://github.com/FerretDB/FerretDB/issues/3399
	return new(backends.CreateIndexesResult), nil
}

// DropIndexes implements backends.Collection interface.
func (c *collection) DropIndexes(ctx context.Context, params *backends.DropIndexesParams) (*backends.DropIndexesResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3397
	return new(backends.DropIndexesResult), nil
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)
