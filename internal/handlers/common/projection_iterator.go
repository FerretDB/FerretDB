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
	"errors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// ProjectionIterator returns an iterator that projects documents returned by the underlying iterator.
//
// Next method returns the next projected document.
//
// Close method closes the underlying iterator.
// For that reason, there is no need to track both iterators.
func ProjectionIterator(iter types.DocumentsIterator, projection *types.Document) (types.DocumentsIterator, error) {
	projectionValidated, inclusion, err := validateProjection(projection)
	if errors.Is(err, errProjectionEmpty) {
		return iter, nil
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &projectionIterator{
		iter:       iter,
		projection: projectionValidated,
		inclusion:  inclusion,
	}, nil
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
