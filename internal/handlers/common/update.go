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
	"math/rand"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UpdateDocument updates the given document with a series of update operators.
func UpdateDocument(doc *types.Document, update any) error {
	var updateDoc *types.Document
	var isUpdateArray bool

	switch update := update.(type) {
	case *types.Array:
		updateDoc = must.NotFail(update.Get(0)).(*types.Document)
		isUpdateArray = true
	case *types.Document:
		updateDoc = update
	default:
		return NewError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", update))
	}

	for _, updateOp := range updateDoc.Keys() {
		updateV := must.NotFail(updateDoc.Get(updateOp))

		switch updateOp {
		case "$set":
			setDoc, err := AssertType[*types.Document](updateV)
			if err != nil {
				return err
			}

			for _, setKey := range setDoc.Keys() {
				setValue := must.NotFail(setDoc.Get(setKey))
				if exprs, ok := setValue.(*types.Document); ok && isUpdateArray {
					expr := exprs.Keys()
					if expr[0] == "$rand" {
						setValue = rand.Float64()
					}
				}
				if err = doc.Set(setKey, setValue); err != nil {
					return lazyerrors.Error(err)
				}
			}

		default:
			return NewError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", updateOp))
		}
	}

	return nil
}
