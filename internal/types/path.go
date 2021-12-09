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

package types

import "fmt"

func GetByPath(str any, path ...any) (any, error) {
	if len(path) == 0 {
		return str, nil
	}

	p, path := path[0], path[1:]

	switch str := str.(type) {
	case Array:
		index, ok := p.(int)
		if !ok {
			return nil, fmt.Errorf("types.GetByPath: can't access %[1]T by path %[2]v (%[2]T)", str, p)
		}
		nextStr, err := str.Get(index)
		if err != nil {
			return nil, fmt.Errorf("types.GetByPath: %w", err)
		}
		return GetByPath(nextStr, path...)

	case Document:
		key, ok := p.(string)
		if !ok {
			return nil, fmt.Errorf("types.GetByPath: can't access %[1]T by path %[2]v (%[2]T)", str, p)
		}
		nextStr, err := str.Get(key)
		if err != nil {
			return nil, fmt.Errorf("types.GetByPath: %w", err)
		}
		return GetByPath(nextStr, path...)

	default:
		return nil, fmt.Errorf("types.GetByPath: can't access %[1]T by path %[2]v (%[2]T)", str, p)
	}
}
