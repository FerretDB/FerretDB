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

	"github.com/FerretDB/FerretDB/internal/types"
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
