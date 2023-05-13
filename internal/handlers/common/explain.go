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
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
)

// ExplainParams represents parameters for the explain command.
type ExplainParams struct {
	DB         string          `ferretdb:"$db"`
	Collection string          `ferretdb:"collection"`
	Explain    *types.Document `ferretdb:"explain"`

	Filter *types.Document `ferretdb:"filter,opt"`
	Sort   *types.Document `ferretdb:"sort,opt"`

	StagesDocs []any           `ferretdb:"-"`
	Aggregate  bool            `ferretdb:"-"`
	Command    *types.Document `ferretdb:"-"`

	Verbosity string `ferretdb:"verbosity,ignored"`
}

// GetExplainParams returns the parameters for the explain command.
func GetExplainParams(document *types.Document, l *zap.Logger) (*ExplainParams, error) {
	var params ExplainParams

	err := commonparams.ExtractParams(document, "explain", &params, l)
	if err != nil {
		return nil, err
	}

	//if cmd.Command() == "aggregate" {
	//	var pipeline *types.Array
	//
	//	pipeline, err = GetRequiredParam[*types.Array](explain, "pipeline")
	//	if err != nil {
	//		return nil, commonerrors.NewCommandErrorMsgWithArgument(
	//			commonerrors.ErrMissingField,
	//			"BSON field 'aggregate.pipeline' is missing but a required field",
	//			document.Command(),
	//		)
	//	}
	//
	//	stagesDocs = must.NotFail(iterator.ConsumeValues(pipeline.Iterator()))
	//	for _, d := range stagesDocs {
	//		if _, ok := d.(*types.Document); !ok {
	//			return nil, commonerrors.NewCommandErrorMsgWithArgument(
	//				commonerrors.ErrTypeMismatch,
	//				"Each element of the 'pipeline' array must be an object",
	//				document.Command(),
	//			)
	//		}
	//	}
	//}

	return &params, nil
}
