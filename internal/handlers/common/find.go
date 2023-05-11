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
	"fmt"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
)

// FindParams contains `find` command parameters supported by at least one handler.
//
//nolint:vet // for readability
type FindParams struct {
	DB          string          `ferretdb:"$db"`
	Collection  string          `ferretdb:"collection"`
	Filter      *types.Document `ferretdb:"filter,opt"`
	Sort        *types.Document `ferretdb:"sort,opt"`
	Projection  *types.Document `ferretdb:"projection,opt"`
	Skip        int64           `ferretdb:"skip,opt"`
	Limit       int64           `ferretdb:"limit,opt"`
	BatchSize   int64           `ferretdb:"batchSize,opt"`
	SingleBatch bool            `ferretdb:"singleBatch,opt"`
	Comment     string          `ferretdb:"comment,opt"`
	MaxTimeMS   int32           `ferretdb:"maxTimeMS,opt"`

	Collation any `ferretdb:"collation,unimplemented"`
	Let       any `ferretdb:"let,unimplemented"`

	AllowDiskUse any `ferretdb:"allowDiskUse,ignored"`
	ReadConcern  any `ferretdb:"readConcern,ignored"`
	Max          any `ferretdb:"max,ignored"`
	Min          any `ferretdb:"min,ignored"`
	Hint         any `ferretdb:"hint,ignored"`

	ReturnKey           any `ferretdb:"returnKey,non-default"`
	ShowRecordId        any `ferretdb:"showRecordId,non-default"`
	Tailable            any `ferretdb:"tailable,non-default"`
	OplogReplay         any `ferretdb:"oplogReplay,non-default"`
	NoCursorTimeout     any `ferretdb:"noCursorTimeout,non-default"`
	AwaitData           any `ferretdb:"awaitData,non-default"`
	AllowPartialResults any `ferretdb:"allowPartialResults,non-default"`
}

// GetFindParams returns `find` command parameters.
func GetFindParams(doc *types.Document, l *zap.Logger) (*FindParams, error) {
	var params FindParams

	err := commonparams.ExtractParams(doc, "find", &params, l)
	if err != nil {
		return nil, err
	}

	// TODO: process default values somehow
	params.BatchSize = 101

	return &params, nil
}

func getFindparams(doc *types.Document, l *zap.Logger) (*FindParams, error) {
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
			fmt.Sprintf("collection name has invalid type %s", commonparams.AliasFromType(collection)),
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

	if s, _ := doc.Get("skip"); s != nil {
		if res.Skip, err = commonparams.GetWholeParamStrict("find", "skip", s); err != nil {
			return nil, err
		}
	}

	if l, _ := doc.Get("limit"); l != nil {
		if res.Limit, err = GetLimitParam("find", l); err != nil {
			return nil, err
		}
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/2005

	//if res.BatchSize, err = GetOptionalParam(doc, "batchSize", int32(101)); err != nil {
	//	return nil, err
	//}

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

	Ignored(doc, l, "allowDiskUse", "hint", "max", "min", "readConcern")

	Unimplemented(doc, "collation", "let")

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

	return &res, nil
}
