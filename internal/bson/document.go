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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
)

// field represents a single Document field in the (partially) decoded form.
type field struct {
	value any
	name  string
}

// Document represents a BSON document a.k.a object in the (partially) decoded form.
//
// It may contain duplicate field names.
type Document struct {
	*wirebson.Document // embed to delegate method
}

// TypesDocumentFromOpMsg gets a raw document, decodes, converts to [*types.Document]
// and validates it.
func TypesDocumentFromOpMsg(msg *wire.OpMsg) (*types.Document, error) {
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
		return nil, newValidationError(fmt.Errorf("bson.TypesDocumentFromOpMsg: validation failed for %v with: %v",
			types.FormatAnyValue(tDoc),
			err,
		))
	}

	return tDoc, nil
}

// TypesDocument gets a document, decodes and converts to [*types.Document].
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

// NewDocument creates a new Document from the given pairs of field names and values.
func NewDocument(pairs ...any) (*Document, error) {
	d, err := wirebson.NewDocument(pairs...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &Document{Document: d}, nil
}

// MakeDocument creates a new empty Document with the given capacity.
func MakeDocument(cap int) *Document {
	return &Document{
		Document: wirebson.MakeDocument(cap),
	}
}

// Freeze prevents document from further field modifications.
// Any methods that would modify document fields will panic.
//
// It is safe to call Freeze multiple times.
func (doc *Document) Freeze() {
	doc.Document.Freeze()
}

// FieldNames returns a slice of field names in the Document.
//
// If document contains duplicate field names, the same name may appear multiple times.
func (doc *Document) FieldNames() []string {
	return doc.Document.FieldNames()
}

// Get returns a value of the field with the given name.
//
// It returns nil if the field is not found.
// If document contains duplicate field names, it returns the first one.
//
// TODO https://github.com/FerretDB/FerretDB/issues/4208
func (doc *Document) Get(name string) any {
	return doc.Document.Get(name)
}

// Add adds a new field to the Document.
func (doc *Document) Add(name string, value any) error {
	return doc.Add(name, value)
}

// Remove removes the first existing field with the given name.
// It does nothing if the field with that name does not exist.
func (doc *Document) Remove(name string) {
	doc.Document.Remove(name)
}

// Replace sets the value for the first existing field with the given name.
// It does nothing if the field with that name does not exist.
func (doc *Document) Replace(name string, value any) error {
	return doc.Document.Replace(name, value)
}

// Command returns the first field name. This is often used as a command name.
// It returns an empty string if document is nil or empty.
func (doc *Document) Command() string {
	return doc.Document.Command()
}
