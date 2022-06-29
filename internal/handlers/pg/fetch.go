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

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// sqlParam represents options/parameters used for sql query.
type sqlParam struct {
	db         string
	collection string
	comment    string
}

// fetch fetches all documents from the given database and collection.
// If collection doesn't exist it returns an empty slice and no error.
//
// TODO https://github.com/FerretDB/FerretDB/issues/372
func (h *Handler) fetch(ctx context.Context, param sqlParam) ([]*types.Document, error) {
	// Special case: check if collection exists at all
	collectionExists, err := h.pgPool.CollectionExists(ctx, param.db, param.collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	if !collectionExists {
		h.l.Info(
			"Collection doesn't exist, handling a case to deal with a non-existing collection.",
			zap.String("schema", param.db), zap.String("table", param.collection),
		)
		return []*types.Document{}, nil
	}

	res, err := h.pgPool.QueryDocuments(ctx, param.db, param.collection, param.comment)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}
