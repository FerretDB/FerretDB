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
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgGetParameter returns parameter details.
func MsgGetParameter(ctx context.Context, msg *wire.OpMsg, l *zap.Logger) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	showDetails, allParameters, err := extractParam(document)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resDB := must.NotFail(types.NewDocument(
		"authenticationMechanisms", must.NotFail(types.NewDocument(
			"value", must.NotFail(types.NewArray("PLAIN")),
			"settableAtRuntime", false,
			"settableAtStartup", true,
		)),
		"authSchemaVersion", must.NotFail(types.NewDocument(
			"value", int32(5),
			"settableAtRuntime", true,
			"settableAtStartup", true,
		)),
		"featureCompatibilityVersion", must.NotFail(types.NewDocument(
			"value", must.NotFail(types.NewDocument("version", "6.0")),
			"settableAtRuntime", false,
			"settableAtStartup", false,
		)),
		"quiet", must.NotFail(types.NewDocument(
			"value", false,
			"settableAtRuntime", true,
			"settableAtStartup", true,
		)),
		"ok", float64(1),
	))

	resDoc := resDB
	if !showDetails || !allParameters {
		resDoc, err = selectUnit(document, resDB, showDetails, allParameters)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{resDoc},
	}))

	Ignored(document, l, "comment")

	if resDoc.Len() < 2 {
		return &reply, commonerrors.NewCommandErrorMsg(commonerrors.ErrorCode(0), "no option found to get")
	}

	return &reply, nil
}

// selectUnit is makes a selection of requested parameters.
func selectUnit(document, resDB *types.Document, showDetails, allParameters bool) (doc *types.Document, err error) {
	doc = must.NotFail(types.NewDocument())

	iter := resDB.Iterator()
	defer iter.Close()

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, err
		}

		if k == "getParameter" || k == "comment" || k == "$db" {
			continue
		}

		if !allParameters && !document.Has(k) {
			continue
		}

		if !showDetails {
			if itm, ok := v.(*types.Document); ok {
				val, err := itm.Get("value")
				if err != nil {
					continue
				}
				v = val
			}
		}

		doc.Set(k, v)
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
		if v, _ := param.Get("showDetails"); v != nil {
			showDetails, err = commonparams.GetBoolOptionalParam("showDetails", v)
			if err != nil {
				return false, false, lazyerrors.Error(err)
			}
		}

		if v, _ := param.Get("allParameters"); v != nil {
			allParameters, err = commonparams.GetBoolOptionalParam("allParameters", v)
			if err != nil {
				return false, false, lazyerrors.Error(err)
			}
		}
	}

	if getPrm == "*" {
		allParameters = true
	}

	return showDetails, allParameters, nil
}
