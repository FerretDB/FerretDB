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

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
)

// sqlParam represents options/parameters used for sql query.
type sqlParam struct {
	db         string
	collection string
	comment    string
}

// fetch fetches all documents from the given database and collection.
//
// If an error occurs before the fetching, the error is returned immediately.
// The returned channel is always non-nil.
// The channel is closed when all documents are sent; the caller should always drain the channel.
// If an error occurs during fetching, the last message before closing the channel contains an error.
// Context cancelation is not considered an error.
//
// If the collection doesn't exist, fetch returns a closed channel and no error.
//
func (h *Handler) fetch(ctx context.Context, param sqlParam) (<-chan pgdb.FetchedDocs, error) {
	return h.pgPool.QueryDocuments(ctx, param.db, param.collection, param.comment)
}
