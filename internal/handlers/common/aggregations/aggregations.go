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

// ProcessorStage is a common interface for aggregation stages that process document iterator.
type ProcessorStage interface {
	// Process applies an aggregate stage on documents from iterator.
	Process(ctx context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error)
}

// ProducerStage is a common interface for aggregation stages that produce document iterator.
type ProducerStage interface {
	// Produce returns an iterator.
	Produce(ctx context.Context, closer *iterator.MultiCloser) (types.DocumentsIterator, error)
}
