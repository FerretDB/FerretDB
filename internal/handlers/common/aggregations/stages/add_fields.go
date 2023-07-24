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

// addFields represents $addFields stage.
//
//	{ $addFields: { <newField>: <expression>, ... } }
type addFields struct {
	newField *types.Document
	query    aggregations.AggregateQuery
}

// newAddFields validates stage document and creates a new $addFields stage.
func newAddFields(params newStageParams) (aggregations.Stage, error) {
	fields, err := params.stage.Get("$addFields")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	fieldsDoc, ok := fields.(*types.Document)
	if !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrSetBadExpression,
			fmt.Sprintf("$addFields specification stage must be an object, got %T", fields),
			"$addFields (stage)",
		)
	}

	if err := validateFieldPath("$addFields", fieldsDoc); err != nil {
		return nil, err
	}

	if err := validateExpression("$addFields", fieldsDoc); err != nil {
		return nil, err
	}

	return &addFields{
		newField: fieldsDoc,
		query:    params.query,
	}, nil
}

// FetchDocuments implements Stage interface.
func (s *addFields) FetchDocuments(ctx context.Context, closer *iterator.MultiCloser) (types.DocumentsIterator, error) {
	return s.query.QueryDocuments(ctx, closer)
}

// Process implements Stage interface.
func (s *addFields) Process(_ context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error) { //nolint:lll // for readability
	return common.AddFieldsIterator(iter, closer, s.newField), nil
}

// check interfaces
var (
	_ aggregations.Stage = (*addFields)(nil)
)
