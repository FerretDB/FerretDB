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

func ProjectionIterator(iter types.DocumentsIterator, projection *types.Document) (types.DocumentsIterator, error) {
	inclusion, err := isProjectionInclusion(projection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &projectionIterator{
		iter:       iter,
		projection: projection,
		inclusion:  inclusion,
	}, nil
}

type projectionIterator struct {
	iter       types.DocumentsIterator
	projection *types.Document
	inclusion  bool
}

func (iter *projectionIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	_, doc, err := iter.iter.Next()
	if err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	err = projectDocument(iter.inclusion, doc, iter.projection)
	if err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	return unused, doc, nil
}

func (iter *projectionIterator) Close() {
	iter.iter.Close()
}

// check interfaces
var (
	_ types.DocumentsIterator = (*projectionIterator)(nil)
)
