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
	"github.com/FerretDB/FerretDB/internal/types"
)

// ListIndexesParams contains `listIndexes` command parameters.
type ListIndexesParams struct {
	DB         string
	Collection string
}

// GetListIndexesParams returns `listIndexes` command parameters.
func GetListIndexesParams(document *types.Document, l *zap.Logger) (*ListIndexesParams, error) {
	Ignored(document, l, "comment", "cursor")

	var err error
	var params ListIndexesParams

	if params.DB, err = GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	var collectionParam any

	if collectionParam, err = document.Get(document.Command()); err != nil {
		return nil, err
	}

	var ok bool
	if params.Collection, ok = collectionParam.(string); !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	return &params, nil
}
