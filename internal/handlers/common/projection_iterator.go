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
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// ProjectionIterator returns an iterator that projects documents returned by the underlying iterator.
// It will be added to the given closer.
//
// Next method returns the next projected document.
//
// Close method closes the underlying iterator.
func ProjectionIterator(iter types.DocumentsIterator, closer *iterator.MultiCloser, projection *types.Document) (types.DocumentsIterator, error) { //nolint:lll // for readability
	inclusion, err := isProjectionInclusion(projection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := &projectionIterator{
		iter:       iter,
		projection: projectionValidated,
		inclusion:  inclusion,
	}
	closer.Add(res)

	return res, nil
}

// projectionIterator is returned by ProjectionIterator.
type projectionIterator struct {
	iter       types.DocumentsIterator
	projection *types.Document
	inclusion  bool
}

// Next implements iterator.Interface. See ProjectionIterator for details.
func (iter *projectionIterator) Next() (struct{}, *types.Document, error) {
	var unused struct{}

	_, doc, err := iter.iter.Next()
	if err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	projected, err := projectDocument(doc, iter.projection, iter.inclusion)
	if err != nil {
		return unused, nil, lazyerrors.Error(err)
	}

	return unused, projected, nil
}

// Close implements iterator.Interface. See ProjectionIterator for details.
func (iter *projectionIterator) Close() {
	iter.iter.Close()
}

// check interfaces
var (
	_ types.DocumentsIterator = (*projectionIterator)(nil)
)
