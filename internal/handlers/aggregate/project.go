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

package aggregate

import (
	"fmt"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type ProjectedField struct {
	name  string
	alias string
}

func NewField(name, alias string) ProjectedField {
	return ProjectedField{name, alias}
}

func TypeFor(v int32) string {
	if v == 0 {
		return "exclusion"
	} else {
		return "inclusion"
	}
}

func CheckProjectionType(projection *types.Document) (string, error) {
	key := projection.Keys()[0]

	value, ok := must.NotFail(projection.Get(key)).(int32)
	if !ok {
		// FIXME support other projection types, with fixed fields
		return "", fmt.Errorf("invalid projection type: %v", value)
	}

	type_ := TypeFor(value)

	for _, key := range projection.Keys() {
		switch v := must.NotFail(projection.Get(key)).(type) {
		case int32:
			if TypeFor(v) != type_ {
				//lint:ignore ST1005 capitalized to keep compatibility with MongoDB error
				return "", fmt.Errorf("Invalid $project :: caused by :: Cannot do %s on field %s in exclusion projection", type_, key)
			}

		default:
			return "", fmt.Errorf("invalid projection type: %v", v)
		}
	}

	return type_, nil
}

func ParseProjectStage(stages *[]*Stage, projection *types.Document) (*[]ProjectedField, error) {
	type_, err := CheckProjectionType(projection)
	if err != nil {
		return nil, err
	}

	inc := type_ == "inclusion"
	fields := make([]ProjectedField, 0)
	for _, stage := range *stages {
		for _, field := range stage.fields {
			contains := false
			for _, k := range projection.Keys() {
				if k == field.name {
					contains = true
				}
			}
			if !inc && !contains || inc && contains {
				fields = append(fields, NewField(field.name, field.name))
			}
		}
	}

	return &fields, nil
}
