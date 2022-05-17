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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UpdateDocument updates the given document with a series of update operators.
func UpdateDocument(doc, update *types.Document) (skip bool, err error) {
	skip = true
	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOp {
		case "$set":
			var setDoc *types.Document
			setDoc, err = AssertType[*types.Document](updateV)
			if err != nil {
				return
			}

			if len(setDoc.Keys()) == 0 {
				return
			}

			for _, setKey := range setDoc.Keys() {
				setValue := must.NotFail(setDoc.Get(setKey))
				if doc.Has(setKey) {
					skip = skip && (must.NotFail(doc.Get(setKey)) == setValue)
				}
				if err = doc.Set(setKey, setValue); err != nil {
					skip, err = false, lazyerrors.Error(err)
					return
				}
			}

		default:
			skip, err = false, NewError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", updateOp))
			return
		}
	}

	return
}
