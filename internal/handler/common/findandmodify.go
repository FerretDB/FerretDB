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

	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
)

// FindAndModifyParams represent parameters for the findAndModify command.
//
//nolint:vet // for readability
type FindAndModifyParams struct {
	DB                string          `ferretdb:"$db"`
	Collection        string          `ferretdb:"findAndModify,collection"`
	Comment           string          `ferretdb:"comment,opt"`
	Query             *types.Document `ferretdb:"query,opt"`
	Sort              *types.Document `ferretdb:"sort,opt"`
	UpdateValue       any             `ferretdb:"update,opt"`
	Remove            bool            `ferretdb:"remove,opt"`
	Upsert            bool            `ferretdb:"upsert,opt"`
	ReturnNewDocument bool            `ferretdb:"new,opt,numericBool"`
	MaxTimeMS         int64           `ferretdb:"maxTimeMS,opt,wholePositiveNumber"`

	Update      *types.Document `ferretdb:"-"`
	Aggregation *types.Array    `ferretdb:"-"`

	HasUpdateOperators bool `ferretdb:"-"`

	Let          *types.Document `ferretdb:"let,unimplemented"`
	Collation    *types.Document `ferretdb:"collation,unimplemented"`
	Fields       *types.Document `ferretdb:"fields,unimplemented"`
	ArrayFilters *types.Array    `ferretdb:"arrayFilters,unimplemented"`

	Hint                     string          `ferretdb:"hint,ignored"`
	WriteConcern             *types.Document `ferretdb:"writeConcern,ignored"`
	BypassDocumentValidation bool            `ferretdb:"bypassDocumentValidation,ignored"`
	LSID                     any             `ferretdb:"lsid,ignored"`
	TxnNumber                int64           `ferretdb:"txnNumber,ignored"`
	ClusterTime              any             `ferretdb:"$clusterTime,ignored"`
	ReadPreference           *types.Document `ferretdb:"$readPreference,ignored"`
}

// GetFindAndModifyParams returns `findAndModifyParams` command parameters.
func GetFindAndModifyParams(doc *types.Document, l *zap.Logger) (*FindAndModifyParams, error) {
	var params FindAndModifyParams

	err := handlerparams.ExtractParams(doc, "findAndModify", &params, l)
	if err != nil {
		return nil, err
	}

	if params.Collection == "" {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrInvalidNamespace,
			fmt.Sprintf("Invalid namespace specified '%s.'", params.DB),
		)
	}

	if params.UpdateValue == nil && !params.Remove {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrFailedToParse,
			"Either an update or remove=true must be specified",
		)
	}

	if params.ReturnNewDocument && params.Remove {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrFailedToParse,
			"Cannot specify both new=true and remove=true; 'remove' always returns the deleted document",
		)
	}

	if params.UpdateValue != nil {
		switch updateParam := params.UpdateValue.(type) {
		case *types.Document:
			params.Update = updateParam
		case *types.Array:
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrNotImplemented,
				"Aggregation pipelines are not supported yet",
				"update",
			)
		default:
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrFailedToParse,
				"Update argument must be either an object or an array",
				"update",
			)
		}
	}

	if params.Update != nil && params.Remove {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrFailedToParse,
			"Cannot specify both an update and remove=true",
		)
	}

	if params.Upsert && params.Remove {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrFailedToParse,
			"Cannot specify both upsert=true and remove=true",
		)
	}

	hasUpdateOperators, err := HasSupportedUpdateModifiers("findAndModify", params.Update)
	if err != nil {
		return nil, err
	}

	params.HasUpdateOperators = hasUpdateOperators

	return &params, nil
}
