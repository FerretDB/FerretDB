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
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
)

// getJSONSchema returns a marshaled JSON schema received from validator -> $jsonSchema.
func getJSONSchema(tdb *tigrisdb.TigrisDB, doc *types.Document, db string, collection string) (*tjson.Schema, error) {
	collection = tigrisdb.EncodeCollName(collection)

	v, err := common.GetOptionalParam[*types.Document](doc, "validator", types.MakeDocument(0))
	if err != nil {
		return nil, err
	}

	schema, err := common.GetRequiredParam[string](v, "$tigrisSchemaString")
	if err != nil {
		tSchema, err1 := tdb.RefreshCollectionSchema(context.TODO(), db, collection)
		if err1 != nil {
			return nil, err1
		}

		return tSchema, nil
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
