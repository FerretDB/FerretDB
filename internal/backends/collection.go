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

package backends

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/types"
)

type Collection interface {
	Insert(context.Context, *InsertParams) (*InsertResult, error)
}

func CollectionContract(c Collection) Collection {
	return &collectionContract{
		c: c,
	}
}

type collectionContract struct {
	c Collection
}

type InsertParams struct {
	Docs    types.DocumentsIterator
	Ordered bool
}

type InsertResult struct{}

func (cc *collectionContract) Insert(ctx context.Context, params *InsertParams) (res *InsertResult, err error) {
	// defer checkError(err, ErrCollectionDoesNotExist)
	res, err = cc.c.Insert(ctx, params)
	return
}

// check interfaces
var (
	_ Collection = (*collectionContract)(nil)
)
