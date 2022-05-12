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
func UpdateDocument(doc, update *types.Document) error {
	for _, updateOp := range update.Keys() {
		updateV := must.NotFail(update.Get(updateOp))

		switch updateOp {
		case "$set":
			setDoc, err := AssertType[*types.Document](updateV)
			if err != nil {
				return err
			}

			for _, setKey := range setDoc.Keys() {
				setValue := must.NotFail(setDoc.Get(setKey))
				if err = doc.Set(setKey, setValue); err != nil {
					return lazyerrors.Error(err)
				}
			}

		case "$inc":
			incDoc, err := AssertType[*types.Document](updateV)
			if err != nil {
				return err
			}

			for _, incKey := range incDoc.Keys() {
				incValue := must.NotFail(incDoc.Get(incKey))

				if !doc.Has(incKey) {
					must.NoError(doc.Set(incKey, incValue))
					continue
				}

				docValue := must.NotFail(doc.Get(incKey))
				incremented, err := addNumbers(incValue, docValue)
				if err != nil {
					return err
				}
				must.NoError(doc.Set(incKey, incremented))
			}

		default:
			return NewError(ErrNotImplemented, fmt.Errorf("UpdateDocument: unhandled operation %q", updateOp))
		}
	}

	return nil
}

func addNumbers(v1, v2 any) (any, error) {
	switch v1 := v1.(type) {
	case float64:
		switch v2 := v2.(type) {
		case float64:
			return v1 + v2, nil
		case int32:
			return v1 + float64(v2), nil
		case int64:
			return v1 + float64(v2), nil
		default:
			return nil, fmt.Errorf("bad type")
		}
	case int32:
		switch v2 := v2.(type) {
		case float64:
			return v2 + float64(v1), nil
		case int32:
			return v1 + v2, nil
		case int64:
			return v2 + int64(v1), nil
		default:
			return nil, fmt.Errorf("bad type")
		}
	case int64:
		switch v2 := v2.(type) {
		case float64:
			return v2 + float64(v1), nil
		case int32:
			return v1 + int64(v2), nil
		case int64:
			return v1 + v2, nil
		default:
			return nil, fmt.Errorf("bad type")
		}
	default:
		return nil, fmt.Errorf("bad type")
	}
}
