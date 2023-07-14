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
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// set represents $set stage.
//
// The $set stage is an alias for $addFields.
//
//	{ $set: { <newField>: <expression>, ... } }
type set struct {
	newField *types.Document
}

// newSet validates stage document and creates a new $set stage.
func newSet(stage *types.Document) (aggregations.Stage, error) {
	fields, err := stage.Get("$set")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	fieldsDoc, ok := fields.(*types.Document)
	if !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrSetBadExpression,
			fmt.Sprintf("$set specification stage must be an object, got %T", fields),
			"$set (stage)",
		)
	}

	if err := validateFieldPath("$set", fieldsDoc); err != nil {
		return nil, err
	}

	if err := validateExpression("$set", fieldsDoc); err != nil {
		return nil, err
	}

	return &set{
		newField: fieldsDoc,
	}, nil
}

// Process implements Stage interface.
func (s *set) Process(_ context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error) { //nolint:lll // for readability
	return common.AddFieldsIterator(iter, closer, s.newField), nil
}

// check interfaces
var (
	_ aggregations.Stage = (*set)(nil)
)
