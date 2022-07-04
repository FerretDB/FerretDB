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

func AddSortStage(stages *[]*Stage, sort *types.Document) error {
	stage := NewEmptyStage()
	if len(*stages) > 0 {
		stage = *(*stages)[len(*stages)-1]
	}
	*stages = append(*stages, &stage)

	for _, field := range sort.Keys() {
		value := must.NotFail(sort.Get(field))
		intVal, ok := value.(int32)
		if !ok || intVal != 1 && intVal != -1 {
			return fmt.Errorf("invalid sort order: %v", value)
		}

		if intVal == 1 {
			stage.AddSortField(field, SortAsc)
		} else {
			stage.AddSortField(field, SortDesc)
		}
	}

	return nil
}
