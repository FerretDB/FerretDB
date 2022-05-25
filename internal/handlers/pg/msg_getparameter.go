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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetParameter OpMsg used to get parameter.
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

	common.Ignored(document, h.l, "comment")

	if resDoc.Len() < 2 {
		return &reply, common.NewErrorMsg(common.ErrorCode(0), "no option found to get")
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
		err = doc.Set(k, item)
		if err != nil {
			return nil, err
		}
	}

	if doc.Len() < 1 {
		err := doc.Set("ok", float64(0))
		if err != nil {
			return nil, err
		}
		return doc, nil
	}

	err = doc.Set("ok", float64(1))
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// extractParam is getting parameters showDetails & allParameters from the request.
func extractParam(document *types.Document) (showDetails, allParameters bool, err error) {
	getPrm, err := document.Get("getParameter")
	if err != nil {
		return false, false, lazyerrors.Error(err)
	}

	if param, ok := getPrm.(*types.Document); ok {
		var errType string
		show, _ := param.Get("showDetails")
		showDetails, errType = convertBool(show)
		if errType != "" {
			s := fmt.Sprintf(`BSON field 'getParameter.showDetails' is the wrong type '%s', `+
				`expected types '[bool, long, int, decimal, double']`, errType,
			)
			return false, false, common.NewErrorMsg(common.ErrTypeMismatch, s)
		}

		all, _ := param.Get("allParameters")
		allParameters, errType = convertBool(all)
		if errType != "" {
			s := fmt.Sprintf(`BSON field 'getParameter.allParameters' is the wrong type '%s', `+
				`expected types '[bool, long, int, decimal, double']`, errType,
			)
			return false, false, common.NewErrorMsg(common.ErrTypeMismatch, s)
		}
	} else {
		if getPrm == "*" {
			allParameters = true
		}
	}
	return showDetails, allParameters, nil
}

// convertBool converts numeric types to bool type.
func convertBool(v any) (val bool, errType string) {
	if v == nil {
		return false, ""
	}

	switch v := v.(type) {
	case types.NullType:
		return false, ""
	case bool:
		return v, ""
	case int32:
		return v != 0, ""
	case int64:
		return v != 0, ""
	case float64:
		return v != 0, ""
	default:
		return false, fmt.Sprintf("%+T", v)
	}
}
