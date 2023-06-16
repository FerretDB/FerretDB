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
	"strings"

	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations/stages/projection"
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
	exclusion *types.Document
}

// newUnset validates unset document and creates a new $unset stage.
func newUnset(stage *types.Document) (aggregations.Stage, error) {
	fields := must.NotFail(stage.Get("$unset"))

	// exclusion contains keys with `false` values to specify projection exclusion later.
	exclusion := must.NotFail(types.NewDocument())

	switch fields := fields.(type) {
	case *types.Array:
		if fields.Len() == 0 {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageUnsetNoPath,
				"$unset specification must be a string or an array with at least one field",
				"$unset (stage)",
			)
		}

		iter := fields.Iterator()
		defer iter.Close()

		conflictFields := map[string]struct{}{}
		unset := struct{}{}

		for {
			_, v, err := iter.Next()
			if err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}

				return nil, lazyerrors.Error(err)
			}

			field, ok := v.(string)
			if !ok {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrStageUnsetArrElementInvalidType,
					"$unset specification must be a string or an array containing only string values",
					"$unset (stage)",
				)
			}

			path, err := validateUnsetField(field)
			if err != nil {
				return nil, err
			}

			if _, ok = conflictFields[field]; ok {
				return nil, commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrUnsetPathDuplicate,
					fmt.Sprintf("Invalid $unset :: caused by :: Path collision at %s", field),
					"$unset (stage)",
				)
			}

			trimmedPath := *path
			i := path.Len()

			// Add field and its parents fields to conflictFields.
			// For a field "v.foo.bar", conflictFields contains "v", "v.foo" and "v.foo.bar".
			for {
				if _, ok = conflictFields[trimmedPath.String()]; ok {
					return nil, commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrUnsetPathOverlap,
						fmt.Sprintf(
							"Invalid $unset :: caused by :: Path collision at %s remaining portion %s",
							path.String(),
							strings.Join(path.Slice()[i:], "."),
						),
						"$unset (stage)",
					)
				}

				conflictFields[trimmedPath.String()] = unset

				if trimmedPath.Len() <= 1 {
					break
				}

				trimmedPath = trimmedPath.TrimSuffix()
				i--
			}

			exclusion.Set(field, false)
		}
	case string:
		if _, err := validateUnsetField(fields); err != nil {
			return nil, err
		}

		exclusion.Set(fields, false)
	default:
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageUnsetInvalidType,
			"$unset specification must be a string or an array",
			"$unset (stage)",
		)
	}

	return &unset{
		exclusion: exclusion,
	}, nil
}

// Process implements Stage interface.
func (u *unset) Process(_ context.Context, iter types.DocumentsIterator, closer *iterator.MultiCloser) (types.DocumentsIterator, error) { //nolint:lll // for readability
	// Use $project to unset fields, $unset is alias for $project exclusion.
	return projection.ProjectionIterator(iter, closer, u.exclusion)
}

// Type implements Stage interface.
func (u *unset) Type() aggregations.StageType {
	return aggregations.StageTypeDocuments
}

// validateUnsetField returns error on invalid field value.
func validateUnsetField(field string) (*types.Path, error) {
	if field == "" {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrEmptyFieldPath,
			"Invalid $unset :: caused by :: FieldPath cannot be constructed with empty string",
			"$unset (stage)",
		)
	}

	if strings.HasPrefix(field, "$") {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrFieldPathInvalidName,
			"Invalid $unset :: caused by :: FieldPath field names may not start with '$'. "+
				"Consider using $getField or $setField.",
			"$unset (stage)",
		)
	}

	if strings.HasSuffix(field, ".") {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrInvalidFieldPath,
			"Invalid $unset :: caused by :: FieldPath must not end with a '.'.",
			"$unset (stage)",
		)
	}

	path, err := types.NewPathFromString(field)
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrPathContainsEmptyElement,
			"Invalid $unset :: caused by :: FieldPath field names may not be empty strings.",
			"$unset (stage)",
		)
	}

	return &path, nil
}

// check interfaces
var (
	_ aggregations.Stage = (*unset)(nil)
)
