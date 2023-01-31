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

package common

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

// defaultBatchSize is the default batch size for find operation.
const defaultBatchSize = 101

// FindParams contains `find` command parameters supported by at least one handler.
//
//nolint:vet // for readability
type FindParams struct {
	DB         string
	Collection string
	Filter     *types.Document
	Sort       *types.Document
	Projection *types.Document
	Limit      int64
	Comment    string
	MaxTimeMS  int32
	BatchSize  int32
}

// GetFindParams returns `find` command parameters.
func GetFindParams(doc *types.Document, l *zap.Logger) (*FindParams, error) {
	var res FindParams
	var err error

	if res.DB, err = GetRequiredParam[string](doc, "$db"); err != nil {
		return nil, err
	}

	collection, err := doc.Get(doc.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if res.Collection, ok = collection.(string); !ok {
		return nil, NewCommandErrorMsgWithArgument(
			ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", AliasFromType(collection)),
			doc.Command(),
		)
	}

	if res.Filter, err = GetOptionalParam(doc, "filter", res.Filter); err != nil {
		return nil, err
	}

	if res.Sort, err = GetOptionalParam(doc, "sort", res.Sort); err != nil {
		return nil, NewCommandErrorMsgWithArgument(
			ErrTypeMismatch,
			"Expected field sort to be of type object",
			"sort",
		)
	}

	if res.Projection, err = GetOptionalParam(doc, "projection", res.Projection); err != nil {
		return nil, err
	}

	Ignored(doc, l, "hint")

	if err = UnimplementedNonDefault(doc, "skip", func(v any) bool {
		n, e := GetWholeNumberParam(v)
		return e == nil && n == 0
	}); err != nil {
		return nil, err
	}

	if l, _ := doc.Get("limit"); l != nil {
		if res.Limit, err = GetWholeNumberParam(l); err != nil {
			return nil, err
		}
	}

	Ignored(doc, l, "singleBatch")

	if res.BatchSize, err = GetOptionalParam(doc, "batchSize", int32(defaultBatchSize)); err != nil {
		return nil, err
	}

	if res.BatchSize < 0 {
		return nil, NewCommandError(commonerrors.ErrBatchSizeNegative,
			fmt.Errorf("BSON field 'batchSize' value must be >= 0, actual value '%d'", res.BatchSize),
		)
	}

	if res.BatchSize < defaultBatchSize {
		res.BatchSize = defaultBatchSize
	}

	if res.Comment, err = GetOptionalParam(doc, "comment", res.Comment); err != nil {
		return nil, err
	}

	if res.MaxTimeMS, err = GetOptionalPositiveNumber(doc, "maxTimeMS"); err != nil {
		return nil, err
	}

	Ignored(doc, l, "readConcern")

	Ignored(doc, l, "max", "min")

	// boolean flags that default to false
	for _, k := range []string{
		"returnKey",
		"showRecordId",
		"tailable",
		"awaitData",
		"oplogReplay",
		"noCursorTimeout",
		"allowPartialResults",
	} {
		if err = UnimplementedNonDefault(doc, k, func(v any) bool {
			b, ok := v.(bool)
			return ok && !b
		}); err != nil {
			return nil, err
		}
	}

	Unimplemented(doc, "collation")

	Ignored(doc, l, "allowDiskUse")

	Unimplemented(doc, "let")

	return &res, nil
}

// MakeFindReplyParameters returns `find` command reply parameters.
// If the amount of documents is more than the batch size, the rest of the documents will be saved in the cursor.
func MakeFindReplyParameters(
	ctx context.Context,
	resDocs []*types.Document, batch int,
	p iterator.Interface[int, *types.Document],
	tx pgx.Tx,
	filter *types.Document,
) (
	*types.Array, int64,
) {
	id := int64(0)
	firstBatch := types.MakeArray(len(resDocs))

	if len(resDocs) == 0 {
		return firstBatch, id
	}

	if batch > len(resDocs) {
		batch = len(resDocs)
	}

	for i := 0; i < batch; i++ {
		firstBatch.Append(resDocs[i])
	}

	if p != nil {
		id = conninfo.Get(ctx).SetCursor(tx, p, filter)
	}

	return firstBatch, id
}
