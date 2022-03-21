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

// isProjectionInclusion: projection can be only inclusion or exlusion. Validate and return true if inclusion.
// Exception for the _id field
func isProjectionInclusion(projection *types.Document) (inclusion bool, err error) {
	errMsg := "projection must contain only inclusions or exclusions"
	var exclusion bool
	for k, v := range projection.Map() {
		if k == "_id" { // _id is a special case and can be both
			continue
		}
		switch v := v.(type) {
		case bool:
			if v {
				if exclusion {
					err = lazyerrors.New(errMsg)
					return
				}
				inclusion = true
			} else {
				if inclusion {
					err = lazyerrors.New(errMsg)
					return
				}
				exclusion = true
			}
		case int32, int64, float64:
			if compareScalars(v, int32(0)) == equal {
				if inclusion {
					err = lazyerrors.New(errMsg)
					return
				}
				exclusion = true
			} else {
				if exclusion {
					err = lazyerrors.New(errMsg)
					return
				}
				inclusion = true
			}
		default:
			err = lazyerrors.Errorf("unsupported operation %s %v (%T)", k, v, v)
			return
		}
	}
	return
}

// ProjectDocuments modifies given documents in places according to the given projection.
func ProjectDocuments(docs []*types.Document, projection *types.Document) error {
	if projection.Len() == 0 {
		return nil
	}

	inclusion, err := isProjectionInclusion(projection)
	if err != nil {
		return err
	}

	projectionMap := projection.Map()
	for i := 0; i < len(docs); i++ {
		for k := range docs[i].Map() {

			v, ok := projectionMap[k]
			if !ok {
				if k == "_id" { // if _id is not in projection map, do not do anything with it
					continue
				}
				if inclusion {
					docs[i].Remove(k)
				}
				continue
			}

			switch v := v.(type) { // found in the projection
			case bool:
				if !v {
					docs[i].Remove(k)
				}
			case int32, int64, float64:
				if compareScalars(v, int32(0)) == equal {
					docs[i].Remove(k)
				}
			default:
				return lazyerrors.Errorf("unsupported operation %s %v (%T)", k, v, v)
			}
		}
	}

	return nil
}
