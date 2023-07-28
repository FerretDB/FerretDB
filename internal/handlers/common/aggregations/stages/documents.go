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

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// DocumentsDataSource queries and fetches all documents.
type DocumentsDataSource interface {
	Query(ctx context.Context, closer *iterator.MultiCloser) (types.DocumentsIterator, error)
}

// documents contains datasource to fetch document from.
type documents struct {
	datasource DocumentsDataSource
}

// NewDocumentProducer creates a new document producer stage.
func NewDocumentProducer(datasource DocumentsDataSource) (aggregations.ProducerStage, error) {
	return &documents{
		datasource: datasource,
	}, nil
}

// Produce implements ProducerStage interface.
func (d *documents) Produce(ctx context.Context, closer *iterator.MultiCloser) (types.DocumentsIterator, error) { //nolint:lll // for readability
	return d.datasource.Query(ctx, closer)
}

// check interfaces
var (
	_ aggregations.ProducerStage = (*documents)(nil)
)
