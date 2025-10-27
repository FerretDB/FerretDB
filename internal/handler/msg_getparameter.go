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

package handler

import (
	"context"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// msgGetParameter implements `getParameter` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgGetParameter(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	getParameter := doc.Get(doc.Command())

	showDetails, allParameters, err := extractGetParameter(getParameter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	parameters := wirebson.MustDocument(
		"authenticationMechanisms", wirebson.MustDocument(
			"value", wirebson.MustArray("SCRAM-SHA-1", "SCRAM-SHA-256"),
			"settableAtRuntime", false,
			"settableAtStartup", true,
		),
		"authSchemaVersion", wirebson.MustDocument(
			"value", int32(5),
			"settableAtRuntime", true,
			"settableAtStartup", true,
		),
		"featureCompatibilityVersion", wirebson.MustDocument(
			// TODO https://github.com/FerretDB/FerretDB/issues/5073
			"value", wirebson.MustDocument("version", "7.0"),
			"settableAtRuntime", false,
			"settableAtStartup", false,
		),
		"quiet", wirebson.MustDocument(
			"value", false,
			"settableAtRuntime", true,
			"settableAtStartup", true,
		),
		// parameters are alphabetically ordered
	)

	res, err := selectParameters(doc, parameters, showDetails, allParameters)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if len(res.FieldNames()) < 1 {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrInvalidOptions,
			"no option found to get",
			doc.Command(),
		)
	}

	must.NoError(res.Add("ok", float64(1)))

	return middleware.ResponseDoc(req, res)
}

// selectParameters makes a selection of requested parameters.
func selectParameters(document, parameters *wirebson.Document, showDetails, allParameters bool) (*wirebson.Document, error) {
	params := parameters.FieldNames()
	resDoc := wirebson.MakeDocument(len(params))

	for _, k := range params {
		if !allParameters && document.Get(k) == nil {
			continue
		}

		v := parameters.Get(k)

		if !showDetails {
			doc, err := v.(wirebson.AnyDocument).Decode()
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			v = doc.Get("value")
		}

		must.NoError(resDoc.Add(k, v))
	}

	return resDoc, nil
}

// extractGetParameter retrieves showDetails & allParameters options set on the getParameter value.
func extractGetParameter(getParameter any) (showDetails, allParameters bool, err error) {
	if getParameter == "*" {
		allParameters = true
		return
	}

	if param, ok := getParameter.(wirebson.AnyDocument); ok {
		var doc *wirebson.Document

		doc, err = param.Decode()
		if err != nil {
			return false, false, lazyerrors.Error(err)
		}

		if v := doc.Get("showDetails"); v != nil {
			showDetails, err = getBoolParam("showDetails", v)
			if err != nil {
				return false, false, lazyerrors.Error(err)
			}
		}

		if v := doc.Get("allParameters"); v != nil {
			allParameters, err = getBoolParam("allParameters", v)
			if err != nil {
				return false, false, lazyerrors.Error(err)
			}
		}
	}

	return showDetails, allParameters, nil
}
