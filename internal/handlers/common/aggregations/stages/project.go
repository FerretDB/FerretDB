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
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
)

// project represents $project stage.
//
//	{
//	  $project: {
//	    <output field1>: <expression1>,
//	    ...
//	    <output fieldN>: <expressionN>
//	  }
type project struct {
	projection *types.Document
	inclusion  bool
}

// newProject validates projection document and creates a new $project stage.
func newProject(stage *types.Document) (Stage, error) {
	fields, err := commonparams.GetRequiredParam[*types.Document](stage, "$project")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrProjectBadExpression,
			"$project specification must be an object",
			"$project (stage)",
		)
	}

	var cmdErr *commonerrors.CommandError

	validated, inclusion, err := common.ValidateProjection(fields)
	if errors.As(err, &cmdErr) {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			cmdErr.Code(),
			fmt.Sprintf("Invalid $project :: caused by :: %s", cmdErr.Unwrap()),
			"$project (stage)",
		)
	}

	if err != nil {
		return nil, err
	}

	return &project{
		projection: validated,
		inclusion:  inclusion,
	}, nil
}

// Process implements Stage interface.
func (p *project) Process(_ context.Context, in []*types.Document) ([]*types.Document, error) {
	var out []*types.Document

	for _, doc := range in {
		projected, err := common.ProjectDocument(doc, p.projection, p.inclusion)
		if err != nil {
			return nil, err
		}

		out = append(out, projected)
	}

	return out, nil
}

// Type implements Stage interface.
func (p *project) Type() StageType {
	return StageTypeDocuments
}
