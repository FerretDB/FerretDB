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

// GetParameter returns parameter details.
func GetParameter(_ context.Context, msg *wire.OpMsg, l *zap.Logger) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	getParameter := must.NotFail(document.Get("getParameter"))

	showDetails, allParameters, err := extractGetParameter(getParameter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	Ignored(document, l, "comment")

	parameters := must.NotFail(types.NewDocument(
		// to add a new parameter, fill template and place it in the alphabetical order position
		//"<name>", must.NotFail(types.NewDocument(
		//	"value", <value>,
		//	"settableAtRuntime", <bool>,
		//	"settableAtStartup", <bool>,
		//)),
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
		// parameters are alphabetically ordered
	))

	resDoc, err := selectParameters(document, parameters, showDetails, allParameters)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if resDoc.Len() < 1 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrorCode(0),
			"no option found to get",
			document.Command(),
		)
	}

	resDoc.Set("ok", float64(1))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{resDoc},
	}))

	return &reply, nil
}

// selectParameters makes a selection of requested parameters.
func selectParameters(document, parameters *types.Document, showDetails, allParameters bool) (resDoc *types.Document, err error) {
	resDoc = must.NotFail(types.NewDocument())

	iter := parameters.Iterator()
	defer iter.Close()

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, lazyerrors.Error(err)
		}

		if !allParameters && !document.Has(k) {
			continue
		}

		if !showDetails {
			v = must.NotFail(v.(*types.Document).Get("value"))
		}

		resDoc.Set(k, v)
	}

	return resDoc, nil
}

// extractGetParameter retrieves showDetails & allParameters options set on the getParameter value.
func extractGetParameter(getParameter any) (showDetails, allParameters bool, err error) {
	if getParameter == "*" {
		allParameters = true
		return
	}

	if param, ok := getParameter.(*types.Document); ok {
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

	return showDetails, allParameters, nil
}
