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

package bson

import (
	"fmt"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Section0Document gets a raw document, decodes, converts to [*types.Document]
// and validates it.
func Section0Document(msg *wire.OpMsg) (*types.Document, error) {
	rDoc, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	tDoc, err := TypesDocument(rDoc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = validateValue(tDoc); err != nil {
		tDoc.Remove("lsid") // to simplify error message

		return nil, newValidationError(fmt.Errorf("bson.Section0Document: validation failed for %v with: %v",
			types.FormatAnyValue(tDoc),
			err,
		))
	}

	return tDoc, nil
}

// AllSectionsDocument first gets the document from section 0, decodes and converts to [*types.Document].
// Then it gets raw documents from sections 1, decodes and append it to the response using section identifier.
// It validates the document.
func AllSectionsDocument(msg *wire.OpMsg) (*types.Document, error) {
	res, err := TypesDocument(msg.RawSection0())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	for _, section := range msg.Sections() {
		if section.Kind == 0 {
			continue
		}

		a := types.MakeArray(len(section.Documents))

		for _, d := range section.Documents {
			var doc *types.Document

			if doc, err = TypesDocument(d); err != nil {
				return nil, lazyerrors.Error(err)
			}

			a.Append(doc)
		}

		res.Set(section.Identifier, a)
	}

	if err = validateValue(res); err != nil {
		res.Remove("lsid") // to simplify error message

		return nil, newValidationError(fmt.Errorf("bson.AllSectionsDocument: validation failed for %v with: %v",
			types.FormatAnyValue(res),
			err,
		))
	}

	return res, nil
}

// NewOpMsg validates the document and converts it to [*wirebson.Document] to create a new OpMsg with it.
func NewOpMsg(doc *types.Document) (*wire.OpMsg, error) {
	if err := validateValue(doc); err != nil {
		doc.Remove("lsid") // to simplify error message

		return nil, newValidationError(fmt.Errorf("bson.NewOpMsg: validation failed for %v with: %v",
			types.FormatAnyValue(doc),
			err,
		))
	}

	return wire.NewOpMsg(must.NotFail(ConvertDocument(doc)))
}

// TypesDocument decodes a document and converts to [*types.Document].
func TypesDocument(doc wirebson.AnyDocument) (*types.Document, error) {
	wDoc, err := doc.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	bDoc := &Document{Document: wDoc}

	tDoc, err := bDoc.Convert()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return tDoc, nil
}
