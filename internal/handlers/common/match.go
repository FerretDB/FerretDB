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
	"log"
	"reflect"

	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// matchDocuments returns true if 2 documents are equal.
//
// TODO move into types.Compare.
func matchDocuments(a, b *types.Document) bool {
	if a == nil {
		log.Panicf("%v is nil", a)
	}
	if b == nil {
		log.Panicf("%v is nil", b)
	}

	keys := a.Keys()
	if !slices.Equal(keys, b.Keys()) {
		return false
	}
	return reflect.DeepEqual(a.Map(), b.Map())
}

// matchArrays returns true if a filter array equals exactly the specified array or
// array contains an element that equals the array.
//
// TODO move into types.Compare.
func matchArrays(filterArr, docArr *types.Array) bool {
	if filterArr == nil {
		log.Panicf("%v is nil", filterArr)
	}
	if docArr == nil {
		log.Panicf("%v is nil", docArr)
	}

	if string(must.NotFail(fjson.Marshal(filterArr))) == string(must.NotFail(fjson.Marshal(docArr))) {
		return true
	}

	for i := 0; i < docArr.Len(); i++ {
		arrValue := must.NotFail(docArr.Get(i))
		if arrValue, ok := arrValue.(*types.Array); ok {
			if string(must.NotFail(fjson.Marshal(filterArr))) == string(must.NotFail(fjson.Marshal(arrValue))) {
				return true
			}
		}
	}

	return false
}
