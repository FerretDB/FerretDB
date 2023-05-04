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

// CountParams represents the parameters for the count command.
type CountParams struct {
	Collation any `name:"collation,unimplemented"`

	Hint        any `name:"hint,ignored"`
	ReadConcern any `name:"readConcern,ignored"`
	Comment     any `name:"comment,ignored"`

	Filter     *types.Document `name:"query"`
	DB         string          `name:"$db"`
	Collection string          `name:"collection"`
	Skip       int64           `name:"skip"`
	Limit      int64           `name:"limit"`
}

// GetCountParams returns the parameters for the count command.
func GetCountParams(document *types.Document, l *zap.Logger) (*CountParams, error) {
	var count CountParams

	err := commonparams.Unmarshal(document, "count", &count, l)
	if err != nil {
		return nil, err
	}

	return &count, nil
}
