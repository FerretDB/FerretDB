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

package aggregations

import (
	"context"
	"fmt"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"strings"
	"time"
)

// newAccumFunc is a type for a function that creates a new group accumulator.
type newAccumFunc func(doc *types.Document) (Stage, error)

// Group is a common interface for accumulation.
type Group interface {
	// Accumulate applies an accumulation on `in` document.
	Process(ctx context.Context, in []*types.Document) ([]*types.Document, error)
}

// accumulators maps all supported group accumulators.
var accumulators = map[string]newAccumFunc{
	// sorted alphabetically
	"$count": newCount,
}

// group represents $group stage.
type group struct {
	fieldExpression any
	accumulators    []accumulator
}

// accumulator represents accumulator of a group.
type accumulator struct {
	field          string
	accumulator    Group
	accumationExpr any
}

// newGroup creates a new $group stage.
func newGroup(stage *types.Document) (Stage, error) {
	fields, err := common.GetRequiredParam[*types.Document](stage, "$group")
	if err != nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupInvalidFields,
			"a group's fields must be specified in an object",
			"$group",
		)
	}

	if fields.Len() == 0 {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupMissingID,
			"a group specification must include an _id",
			"$group",
		)
	}

	var group group
	iter := fields.Iterator()
	defer iter.Close()

	for {
		k, v, err := iter.Next()
		if err == iterator.ErrIteratorDone {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if k == "_id" {
			group.fieldExpression = v
			continue
		}

		accumDoc, ok := v.(*types.Document)
		if !ok || accumDoc.Len() == 0 {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageGroupInvalidField,
				fmt.Sprintf("The field '%s' must be an accumulator object", k),
				"$group",
			)
		}

		if accumDoc.Len() > 1 {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrStageGroupOneAccumulator,
				fmt.Sprintf("The field '%s' must specify one accumulator", k),
				"$group",
			)
		}

		// document contains only one.
		accumK, accumV, err := accumDoc.Iterator().Next()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		newAccumulator, ok := accumulators[accumK]
		if !ok {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrNotImplemented,
				"unimplemented",
				"$group",
			)
		}

		acc, err := newAccumulator(accumDoc)

		a := accumulator{
			field:          k,
			accumulator:    acc,
			accumationExpr: accumV,
		}

		group.accumulators = append(group.accumulators, a)
	}

	if group.fieldExpression == nil {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrStageGroupMissingID,
			"a group specification must include an _id",
			"$group",
		)
	}

	return &group, nil
}

// Process implements Stage interface.
func (g *group) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	var res []*types.Document

	// use fieldExpression to group them
	switch fieldExpression := g.fieldExpression.(type) {
	case string:
		if !strings.HasPrefix(fieldExpression, "$") {
			res = append(res, must.NotFail(types.NewDocument("_id", fieldExpression)))
			break
		}

		key := strings.TrimPrefix(fieldExpression, "$")
		if key == "" {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrGroupInvalidFieldPath,
				"'$' by itself is not a valid FieldPath",
				"$group",
			)
		}

		// find distinct fields
		distinct, err := common.FilterDistinctValues(in, key)
		if err != nil {
			return nil, err
		}

		if distinct.Len() == 0 {
			// return {"_id": null} when no distinct is found
			res = append(res, must.NotFail(types.NewDocument("_id", types.Null)))
			break
		}

		iter := distinct.Iterator()
		if err != nil {
			return nil, err
		}

		defer iter.Close()

		for {
			_, v, err := iter.Next()
			if err == iterator.ErrIteratorDone {
				break
			}

			if err != nil {
				return nil, err
			}

			res = append(res, must.NotFail(types.NewDocument("_id", v)))
		}

	case *types.Document,
		*types.Array,
		float64,
		types.Binary,
		types.ObjectID,
		bool,
		time.Time,
		types.NullType,
		types.Regex,
		int32,
		types.Timestamp,
		int64:
		// use a single document.
		res = append(res, must.NotFail(types.NewDocument("_id", fieldExpression)))

	default:
		panic(fmt.Sprintf("group: unexpected type %[1]T (%#[1]v)", fieldExpression))
	}

	// call accumulator

	return res, nil
}

// check interfaces
var (
	_ Stage = (*group)(nil)
)
