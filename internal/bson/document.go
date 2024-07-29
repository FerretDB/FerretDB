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
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Document represents a BSON document a.k.a object in the (partially) decoded form.
//
// It may contain duplicate field names.
type Document struct {
	*wirebson.Document // embed to delegate method
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
