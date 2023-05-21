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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// unset represents $unset stage.
//
// { $unset: "<field>" }
//
//	or { $unset: [ "<field1>", "<field2>", ... ] }
type unset struct {
	field *types.Document
}

// newUnset validates unset document and creates a new $unset stage.
func newUnset(stage *types.Document) (aggregations.Stage, error) {
	fields, err := stage.Get("$unset")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	fieldsToUnset := must.NotFail(types.NewDocument())

	switch fields := fields.(type) {
	case *types.Array:
		if fields.Len() == 0 {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageUnsetNoPath,
				"no path specified to $unset stage",
				"$unset (stage)",
			)
		}
		fieldIter := fields.Iterator()
		defer fieldIter.Close()
		for {
			_, field, err := fieldIter.Next()
			if err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}
				return nil, lazyerrors.Error(err)
			}
			fieldStr := field.(string)
			fieldsToUnset.Set(fieldStr, false)
		}
	case string:
		if fields == "" {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageUnsetNoPath,
				"no path specified to $unset stage",
				"$unset (stage)",
			)
		}
		fieldsToUnset.Set(fields, false)
	}

	return &unset{
		field: fieldsToUnset,
	}, nil
}

// Process implements Stage interface.

func (u *unset) Process(_ context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error) {
	return common.ProjectionIterator(iter, closer, u.field)
}

// Type implements Stage interface.
func (u *unset) Type() aggregations.StageType {
	return aggregations.StageTypeDocuments
}

// check interfaces
var (
	_ aggregations.Stage = (*unset)(nil)
)
