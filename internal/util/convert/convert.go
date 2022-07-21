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

package convert

import (
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Convert gets bson document and returns *types.Document.
func Convert(m map[string]any) *types.Document {
	doc := new(types.Document)
	for k, mapval := range m {
		must.NoError(doc.Set(k, bson2types(mapval)))
	}
	return doc
}

func bson2types(v any) any {
	switch v := v.(type) {
	case map[string]any:
		m := new(types.Document)
		for k, mapval := range v {
			must.NoError(m.Set(k, bson2types(mapval)))
		}
		return m

	case []any:
		a := new(types.Array)
		for _, arrval := range v {
			must.NoError(a.Append(bson2types(arrval)))
		}
		return a

	case float64,
		string,
		bool,
		time.Time,
		nil,
		int32,
		int64:
		return v
	}
	panic(fmt.Sprintf("unsupported type: %T", v))
}
