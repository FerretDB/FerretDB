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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// ProjectDocuments modifies given documents in places according to the given projection.
func ProjectDocuments(docs []*types.Document, projection *types.Document) error {
	if projection.Len() == 0 {
		return nil
	}
	projectionMap := projection.Map()
	for i := 0; i < len(docs); i++ {
		for k := range docs[i].Map() {
			if k == "_id" {
				continue
			}

			v, ok := projectionMap[k]
			if !ok {
				docs[i].Remove(k)
				continue
			}

			switch v := v.(type) {
			case bool:
				if !v {
					docs[i].Remove(k)
					continue
				}
				continue

			case int32, int64, float32, float64:
				if compareScalars(v, int32(0)) == equal {
					docs[i].Remove(k)
					continue
				}
			default:
				return lazyerrors.Errorf("unsupported operation {%v %v}", k, v)
			}
		}
	}

	return nil
}
