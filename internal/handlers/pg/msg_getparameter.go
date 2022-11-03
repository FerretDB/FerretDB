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

package pg

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetParameter implements HandlerInterface.
func (h *Handler) MsgGetParameter(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	showDetails, allParameters, err := extractParam(document)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resDB := must.NotFail(types.NewDocument(
		"acceptApiVersion2", must.NotFail(types.NewDocument(
			"value", false,
			"settableAtRuntime", true,
			"settableAtStartup", true,
		)),
		"authSchemaVersion", must.NotFail(types.NewDocument(
			"value", int32(5),
			"settableAtRuntime", true,
			"settableAtStartup", true,
		)),
		"tlsMode", must.NotFail(types.NewDocument(
			"value", "disabled",
			"settableAtRuntime", true,
			"settableAtStartup", false,
		)),
		"sslMode", must.NotFail(types.NewDocument(
			"value", "disabled",
			"settableAtRuntime", true,
			"settableAtStartup", false,
		)),
		"quiet", must.NotFail(types.NewDocument(
			"value", false,
			"settableAtRuntime", true,
			"settableAtStartup", true,
		)),
		"ok", float64(1),
	))

	var reply wire.OpMsg
	resDoc := resDB
	if !showDetails || !allParameters {
		resDoc, err = selectUnit(document, resDB, showDetails, allParameters)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	err = reply.SetSections(wire.OpMsgSection{Documents: []*types.Document{resDoc}})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "comment")

	if resDoc.Len() < 2 {
		return &reply, common.NewCommandErrorMsg(common.ErrorCode(0), "no option found to get")
	}

	return &reply, nil
}

// selectUnit is makes a selection of requested parameters.
func selectUnit(document, resDB *types.Document, showDetails, allParameters bool) (doc *types.Document, err error) {
	doc = must.NotFail(types.NewDocument())

	keys := resDB.Keys()
	if !allParameters {
		keys = document.Keys()
	}

	for _, k := range keys {
		if k == "getParameter" || k == "comment" || k == "$db" {
			continue
		}
		item, err := resDB.Get(k)
		if err != nil {
			continue
		}

		if !showDetails {
			if itm, ok := item.(*types.Document); ok {
				val, err := itm.Get("value")
				if err != nil {
					continue
				}
				item = val
			}
		}

		doc.Set(k, item)
	}

	if doc.Len() < 1 {
		doc.Set("ok", float64(0))
		return doc, nil
	}

	doc.Set("ok", float64(1))
	return doc, nil
}

// extractParam is getting parameters showDetails & allParameters from the request.
func extractParam(document *types.Document) (showDetails, allParameters bool, err error) {
	getPrm, err := document.Get("getParameter")
	if err != nil {
		return false, false, lazyerrors.Error(err)
	}

	if param, ok := getPrm.(*types.Document); ok {
		showDetails, err = common.GetBoolOptionalParam(param, "showDetails")
		if err != nil {
			return false, false, lazyerrors.Error(err)
		}
		allParameters, err = common.GetBoolOptionalParam(param, "allParameters")
		if err != nil {
			return false, false, lazyerrors.Error(err)
		}
	}
	if getPrm == "*" {
		allParameters = true
	}

	return showDetails, allParameters, nil
}
