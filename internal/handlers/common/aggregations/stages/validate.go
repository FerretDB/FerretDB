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

package stages

import (
	"errors"
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// validateExpression recursively validates expressions in document and array
// returns error when there is unsupported expression present.
// Currently, it raises ErrNotImplemented if there is any expression.
// Command Errors:
//   - ErrNotImplemented
func validateExpression(stage string, doc *types.Document) error {
	iter := doc.Iterator()
	defer iter.Close()

	for {
		_, v, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			return nil
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		switch value := v.(type) {
		case *types.Document:
			if err := validateExpression(stage, value); err != nil {
				return err
			}
		case *types.Array:
			if err := validateArrayExpression(stage, value); err != nil {
				return err
			}
		}
	}
}

// validateArrayExpression validates each document in array arr and
// returns error when there is unsupported expression present in any document.
// Currently, it raises error if there is any expression.
// Command Errors:
//   - ErrNotImplemented
func validateArrayExpression(stage string, arr *types.Array) error {
	iter := arr.Iterator()
	defer iter.Close()

	for {
		_, v, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			return nil
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		doc, ok := v.(*types.Document)
		if !ok {
			continue
		}

		if err := validateExpression(stage, doc); err != nil {
			return err
		}
	}
}

// validateFieldPath validates each key of fields, it returns error if a field name starts with `$`.
// Command Errors:
//   - ErrFieldPathInvalidName
func validateFieldPath(stage string, fieldsDoc *types.Document) error {
	for _, key := range fieldsDoc.Keys() {
		if strings.HasPrefix(key, "$") {
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFieldPathInvalidName,
				fmt.Sprintf(
					"Invalid %s :: caused by :: FieldPath field names may not start with '$'. "+
						"Consider using $getField or $setField.",
					stage,
				),
				fmt.Sprintf("%s (stage)", stage),
			)
		}
	}

	return nil
}
