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

func AggregateSort(sort *types.Document) (*string, error) {
	order := []string{}

	for _, field := range sort.Keys() {
		value := must.NotFail(sort.Get(field))
		intVal, ok := value.(int32)
		if !ok || intVal != 1 && intVal != -1 {
			return nil, fmt.Errorf("invalid sort order: %v", value)
		}

		if intVal == 1 {
			order = append(order, field)
		} else {
			order = append(order, field+" DESC")
		}
	}

	orderStr := strings.Join(order, ", ")
	return &orderStr, nil
}
