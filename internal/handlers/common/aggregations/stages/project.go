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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
)

type project struct {
	projection *types.Document
	inclusion  bool
}

func newProject(stage *types.Document) (Stage, error) {
	fields, err := common.GetRequiredParam[*types.Document](stage, "$project")
	if err != nil {
		return nil, err
	}

	validated, inclusion, err := common.ValidateProjection(fields)
	if err != nil {
		return nil, err
	}

	return &project{
		projection: validated,
		inclusion:  inclusion,
	}, nil
}

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

func (p *project) Type() StageType {
	return StageTypeDocuments
}
