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

// Package aggregations provides aggregation pipelines.
package aggregations

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// Stage is a common interface for all aggregation stages.
type Stage interface {
	// FetchDocuments fetches document iterator to use for the first stage of the pipeline.
	FetchDocuments(ctx context.Context, closer *iterator.MultiCloser) (types.DocumentsIterator, error)

	// Process applies an aggregate stage on documents from iterator.
	Process(ctx context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error)
}

// AggregateQuery is fetched from database.
type AggregateQuery interface {
	// QueryDocuments fetches documents from the database.
	QueryDocuments(ctx context.Context, closer *iterator.MultiCloser) (types.DocumentsIterator, error)

	// QueryCollStats fetches collection statistics from the database.
	QueryCollStats(ctx context.Context, closer *iterator.MultiCloser) (*CollStats, error)
}

// CollStats describes collection statistics.
type CollStats struct {
	CountObjects   int64
	CountIndexes   int64
	SizeTotal      int64
	SizeIndexes    int64
	SizeCollection int64
}
