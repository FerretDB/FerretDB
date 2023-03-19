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
	"encoding/json"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// getJSONSchema returns a marshaled JSON schema received from validator -> $jsonSchema.
func getJSONSchema(doc *types.Document, collection string) (*tjson.Schema, error) {
	collection = tigrisdb.EncodeCollName(collection)

	v, err := common.GetOptionalParam[*types.Document](doc, "validator", types.MakeDocument(0))
	if err != nil {
		return nil, err
	}

	schema, err := common.GetOptionalParam[string](v, "$tigrisSchemaString", string(getEmptySchema(collection)))
	if err != nil {
		return nil, err
	}

	if schema == "" {
		return nil, common.NewCommandError(common.ErrBadValue, fmt.Errorf("empty schema is not allowed"))
	}

	sch := new(tjson.Schema)
	err = sch.Unmarshal([]byte(schema))

	if err != nil {
		return nil, common.NewCommandError(common.ErrBadValue, err)
	}

	return sch, nil
}

func getEmptySchema(collection string) []byte {
	schema := must.NotFail(tjson.DocumentSchema(must.NotFail(types.NewDocument("_id", types.NewObjectID()))))
	schema.Title = collection

	return must.NotFail(json.Marshal(schema))
}
