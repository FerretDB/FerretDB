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
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// skip represents $skip stage.
type skip struct {
	value int64
}

// newSkip creates a new $skip stage.
func newSkip(stage *types.Document) (Stage, error) {
	value, err := stage.Get("$skip")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	skipValue, err := commonparams.GetSkipStageParam(value)
	if err != nil {
		return nil, err
	}

	return &skip{
		value: skipValue,
	}, nil
}

// Process implements Stage interface.
func (s *skip) Process(ctx context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error) { //nolint:lll // for readability
	return common.SkipIterator(iter, closer, s.value), nil
}

// Type implements Stage interface.
func (s *skip) Type() StageType {
	return StageTypeDocuments
}

// check interfaces
var (
	_ Stage = (*skip)(nil)
)
