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
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type MatchParser struct {
	index int
}

func ParseMatchStage(value interface{}) (*Stage, error) {
	root := NewRootNode()
	mp := MatchParser{}
	err := mp.parse(root, "", "", value)
	if err != nil {
		return nil, err
	}

	stage := NewStage([]string{}, root)
	return &stage, err
}

func (mp *MatchParser) NextIndex() int {
	mp.index++
	return mp.index
}

func (mp *MatchParser) parse(node *FilterNode, key string, field string, value interface{}) error {
	switch key {
	case "$gt":
		node.AddFilter(mp.NextIndex(), field, ">", value)

	case "$gte":
		node.AddFilter(mp.NextIndex(), field, ">=", value)

	case "$lt":
		node.AddFilter(mp.NextIndex(), field, "<", value)

	case "$lte":
		node.AddFilter(mp.NextIndex(), field, "<=", value)

	case "$eq":
		node.AddFilter(mp.NextIndex(), field, "=", value)

	case "$ne":
		node.AddFilter(mp.NextIndex(), field, "<>", value)

	case "$and":
		arr, ok := value.(*types.Array)
		if !ok {
			return NewErrorMsg(ErrBadValue, "$and must be an array")
		}

		opNode := node.AddOp("AND")
		for i := 0; i < arr.Len(); i++ {
			v := must.NotFail(arr.Get(i))
			if err := mp.parse(opNode, "", "", v); err != nil {
				return err
			}
		}

	case "$or":
		arr, ok := value.(*types.Array)
		if !ok {
			return NewErrorMsg(ErrBadValue, "$or must be an array")
		}

		opNode := node.AddOp("OR")
		for i := 0; i < arr.Len(); i++ {
			v := must.NotFail(arr.Get(i))
			if err := mp.parse(opNode, "", "", v); err != nil {
				return err
			}
		}

	case "$not":
		mp.parse(node.AddUnaryOp("NOT"), field, field, value)

	case "$exists":
		if value == false {
			node = node.AddUnaryOp("NOT")
		} else if value != true {
			return NewErrorMsg(ErrBadValue, "$exists must be true or false")
		}
		field, parents := ParseField(field)
		node.AddFilter(mp.NextIndex(), parents, "?", field)

	case "$all":
		arr, ok := value.(*types.Array)
		if !ok {
			return NewErrorMsg(ErrBadValue, "$all must be an array")
		}

		arrVals := []string{}
		for i := 0; i < arr.Len(); i++ {
			arrVals = append(arrVals, fmt.Sprintf("%v", must.NotFail(arr.Get(i))))
		}

		node.AddFilter(mp.NextIndex(), field, "@> (%s)", arrVals)

	case "$in", "$nin":
		arr, ok := value.(*types.Array)
		if !ok {
			return NewErrorMsg(ErrBadValue, key+" must be an array")
		}

		arrVals := []string{}
		for i := 0; i < arr.Len(); i++ {
			arrVals = append(arrVals, fmt.Sprintf("%v", must.NotFail(arr.Get(i))))
		}

		if key == "$nin" {
			node = node.AddUnaryOp("NOT")
		}

		node.AddFilter(mp.NextIndex(), field, "= ANY(%s)", arrVals)

	case "$regex":
		node.AddRawFilter(mp.NextIndex(), field, "~", value)

	default:
		if strings.HasPrefix(key, "$") {
			return NewWriteErrorMsg(
				ErrFailedToParse,
				fmt.Sprintf(
					"Unknown top level operator: %s. Expected a valid aggregate modifier", key,
				),
			)
		}

		switch v := value.(type) {
		case *types.Document:
			for _, k := range v.Keys() {
				value := must.NotFail(v.Get(k))
				err := mp.parse(node, k, key, value)
				if err != nil {
					return err
				}
			}

		default:
			// defaults to equality match
			node.AddFilter(mp.NextIndex(), key, "=", value)
		}
	}

	return nil
}
