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

package tjson

import (
	"encoding/json"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UnmarshalOld should be removed.
func UnmarshalOld(data *driver.Document) (*types.Document, error) {
	var v map[string]any
	err := json.Unmarshal(*data, &v)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = checkUnmarshalSupportedOld(v); err != nil {
		return nil, err
	}

	pairs := make([]any, 2*len(v))
	var i int
	for k, v := range v {
		pairs[i] = k
		pairs[i+1] = v
		i += 2
	}

	doc, err := types.NewDocument(pairs...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// MarshalOld should be removed.
func MarshalOld(v *types.Document) (*driver.Document, error) {
	if v == nil {
		panic("v is nil")
	}

	if err := checkMarshalSupportedOld(v); err != nil {
		return nil, err
	}

	b, err := json.Marshal(v.Map())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	d := driver.Document(b)
	return &d, nil
}

// checkUnmarshalSupportedOld should be removed.
func checkUnmarshalSupportedOld(v any) error {
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case bool, string, float64:
		return nil

	case []any:
		for i := 0; i < len(v); i++ {
			if err := checkUnmarshalSupportedOld(v[i]); err != nil {
				return err
			}
		}
		return nil

	case map[string]any:
		for _, val := range v {
			if err := checkUnmarshalSupportedOld(val); err != nil {
				return err
			}
		}
		return nil

	default:
		return lazyerrors.Errorf("%T not supported", v)
	}
}

// checkMarshalSupportedOld should be removed.
func checkMarshalSupportedOld(v any) error {
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case bool, string, float64:
		return nil

	case *types.Array:
		for i := 0; i < v.Len(); i++ {
			if err := checkMarshalSupportedOld(must.NotFail(v.Get(i))); err != nil {
				return err
			}
		}
		return nil

	case *types.Document:
		m := v.Map()
		for _, val := range m {
			if err := checkMarshalSupportedOld(val); err != nil {
				return err
			}
		}
		return nil

	default:
		return lazyerrors.Errorf("%T not supported", v)
	}
}
