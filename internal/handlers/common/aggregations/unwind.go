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

package aggregations

import (
	"context"
	"errors"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// unwind represents $unwind stage.
type unwind struct {
	field types.Path
}

func newUnwind(stage *types.Document) (Stage, error) {
	field, err := stage.Get("$unwind")
	if err != nil {
		return nil, err
	}

	var path types.Path

	switch field := field.(type) {
	case *types.Document:
		return nil, common.Unimplemented(stage, "$unwind")
	case string:
		path, err = types.NewPathFromString(field)
		switch err {

		}
	}

	return &unwind{
		field: path,
	}, nil
}

func (m *unwind) Process(ctx context.Context, in []*types.Document) ([]*types.Document, error) {
	var out []*types.Document

	for _, doc := range in {
		d, err := doc.GetByPath(m.field)
		if err != nil {
			return nil, err
		}

		switch d := d.(type) {
		//case *types.Document:
		case types.Array:
			iter := d.Iterator()
			defer iter.Close()

			id := must.NotFail(doc.Get("_id"))

			for {
				_, v, err := iter.Next()
				switch {
				case errors.Is(err, iterator.ErrIteratorDone):
					break
				default:
					return nil, err
				}

				newDoc := must.NotFail(types.NewDocument("_id", id, m.field.Suffix(), v))
				out = append(out, newDoc)
			}

		default:
			continue
		}
	}

	return out, nil
}
