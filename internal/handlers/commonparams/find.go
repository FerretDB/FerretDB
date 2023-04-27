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

package commonparams

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
)

// FindParams contains `find` command parameters supported by at least one handler.
//
//nolint:vet // for readability
type FindParams struct {
	DB          string          `name:"$db"`
	Collection  string          `name:"collection"`
	Filter      *types.Document `name:"filter,opt"`
	Sort        *types.Document `name:"sort,opt"`
	Projection  *types.Document `name:"projection,opt"`
	Skip        int64           `name:"skip"`
	Limit       int64           `name:"limit"`
	BatchSize   int32           `name:"batchSize"`
	SingleBatch bool            `name:"singleBatch"`
	Comment     string          `name:"comment"`
	MaxTimeMS   int32           `name:"maxTimeMS"`
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
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", AliasFromType(collection)),
			doc.Command(),
		)
	}

	if res.Filter, err = GetOptionalParam(doc, "filter", res.Filter); err != nil {
		return nil, err
	}

	if res.Sort, err = GetOptionalParam(doc, "sort", res.Sort); err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			"Expected field sort to be of type object",
			"sort",
		)
	}

	if res.Projection, err = GetOptionalParam(doc, "projection", res.Projection); err != nil {
		return nil, err
	}

	Ignored(doc, l, "hint")

	if s, _ := doc.Get("skip"); s != nil {
		if res.Skip, err = GetSkipParam("find", s); err != nil {
			return nil, err
		}
	}

	if l, _ := doc.Get("limit"); l != nil {
		if res.Limit, err = GetLimitParam("find", l); err != nil {
			return nil, err
		}
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/2005

	if res.BatchSize, err = GetOptionalParam(doc, "batchSize", int32(101)); err != nil {
		return nil, err
	}

	if res.BatchSize < 0 {
		return nil, commonerrors.NewCommandError(
			commonerrors.ErrValueNegative,
			fmt.Errorf("BSON field 'batchSize' value must be >= 0, actual value '%d'", res.BatchSize),
		)
	}

	if res.SingleBatch, err = GetOptionalParam(doc, "singleBatch", false); err != nil {
		return nil, err
	}

	if res.Comment, err = GetOptionalParam(doc, "comment", ""); err != nil {
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
		"oplogReplay",
		"noCursorTimeout",
		"awaitData",
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
