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

package pgdb

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/FerretDB/FerretDB/internal/types"

	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/jackc/pgx/v4"
)

const (
	// PostgreSQL max index name length.
	maxIndexNameLength = 63
)

// setMetadataIndex sets the index info in the metadata table.
//
// Indexes are stored in the `indexes` object.
// The FerretDB index name is stored as a key.
// Index settings are stored as a value object:
//   - the corresponding formatted PostgreSQL index name is stored in the pgindex field.
func setIndexMetadata(ctx context.Context, tx pgx.Tx, db, collection, index string) (string, error) {
	var err error

	pgIndex := formatIndexName(index)

	indexMetadata := must.NotFail(types.NewDocument(
		"pgindex", pgIndex,
	))

	addToSetByID(ctx, tx, db, collection, "indexes", index, collection)

	return indexName, nil
}

// formatIndexName returns index name in form <shortened_name>_<name_hash>.
// Changing this logic will break compatibility with existing databases.
func formatIndexName(name string) string {
	hash32 := fnv.New32a()
	_ = must.NotFail(hash32.Write([]byte(name)))

	nameSymbolsLeft := maxIndexNameLength - hash32.Size()*2 - 1
	truncateTo := len(name)

	if truncateTo > nameSymbolsLeft {
		truncateTo = nameSymbolsLeft
	}

	return name[:truncateTo] + "_" + fmt.Sprintf("%x", hash32.Sum([]byte{}))
}
