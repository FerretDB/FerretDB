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
	"errors"
	"fmt"
	"math"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// NewOpMsg validates the document and converts it to [*wirebson.Document] to create a new OpMsg with it.
func NewOpMsg(doc *types.Document) (*wire.OpMsg, error) {
	if err := validateValue(doc); err != nil {
		doc.Remove("lsid") // to simplify error message

		return nil, handlererrors.NewValidationError(fmt.Errorf("wire.OpMsg.Document: validation failed for %v with: %v",
			types.FormatAnyValue(doc),
			err,
		))
	}

	return wire.NewOpMsg(must.NotFail(bson.ConvertDocument(doc)))
}

// OpMsgDocument gets a raw document, decodes and converts to [*types.Document].
// Then it iterates raw documents from sections 1 if any, decodes and append
// them to the response using the section identifier.
// It validates and returns [*types.Document].
func OpMsgDocument(msg *wire.OpMsg) (*types.Document, error) {
	rDoc, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res, err := bson.TypesDocument(rDoc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	for _, section := range msg.Sections() {
		if section.Kind == 0 {
			continue
		}

		docs := section.Documents()
		a := types.MakeArray(len(docs))

		for _, d := range docs {
			var doc *types.Document

			if doc, err = bson.TypesDocument(d); err != nil {
				return nil, lazyerrors.Error(err)
			}

			a.Append(doc)
		}

		res.Set(section.Identifier, a)
	}

	if err = validateValue(res); err != nil {
		res.Remove("lsid") // to simplify error message

		return nil, handlererrors.NewValidationError(fmt.Errorf("wire.OpMsg.Document: validation failed for %v with: %v",
			types.FormatAnyValue(res),
			err,
		))
	}

	return res, nil
}

// validateValue checks given value and returns error if unsupported value was encountered.
func validateValue(v any) error {
	switch v := v.(type) {
	case *types.Document:
		for _, v := range v.Values() {
			if err := validateValue(v); err != nil {
				return err
			}
		}

	case *types.Array:
		for i := 0; i < v.Len(); i++ {
			v := must.NotFail(v.Get(i))
			if err := validateValue(v); err != nil {
				return err
			}
		}

	case float64:
		if math.IsNaN(v) {
			return errors.New("NaN is not supported")
		}
	}

	return nil
}
