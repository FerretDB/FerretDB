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

package tigris

import (
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
)

// getJSONSchema returns a marshaled JSON schema received from validator -> $jsonSchema.
func getJSONSchema(doc *types.Document) (*tjson.Schema, error) {
	v, err := common.GetRequiredParam[*types.Document](doc, "validator")
	if err != nil {
		return nil, err
	}

	schema, err := common.GetRequiredParam[string](v, "$tigrisSchemaString")
	if err != nil {
		return nil, err
	}

	if schema == "" {
		return nil, commonerrors.NewCommandError(commonerrors.ErrBadValue, fmt.Errorf("empty schema is not allowed"))
	}

	sch := new(tjson.Schema)
	err = sch.Unmarshal([]byte(schema))

	if err != nil {
		return nil, commonerrors.NewCommandError(commonerrors.ErrBadValue, err)
	}

	return sch, nil
}
