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

import "github.com/FerretDB/FerretDB/internal/types"

// LimitDocuments returns a subslice of given documents according to the given limit.
func LimitDocuments(docs []*types.Document, limit int64) ([]*types.Document, error) {
	switch {
	case limit == 0:
		return docs, nil
	case limit > 0:
		if int64(len(docs)) <= limit {
			return docs, nil
		}
		return docs[:limit], nil
	default:
		// TODO https://github.com/FerretDB/FerretDB/issues/79
		return nil, NewCommandErrorMsg(ErrNotImplemented, "LimitDocuments: negative limit values are not supported")
	}
}
