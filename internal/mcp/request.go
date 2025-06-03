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

package mcp

import (
	"encoding/json"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// prepareOpMsg creates a new OpMsg from the given pairs of field names and values,
// which can be used as handler command msg.
//
// If any of pair values is nil it's ignored.
func prepareOpMsg(pairs ...any) (*middleware.Request, error) {
	doc, err := prepareDocument(pairs...)
	if err != nil {
		return nil, err
	}

	req, err := wire.NewOpMsg(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &middleware.Request{OpMsg: req}, nil
}

// prepareDocument creates a new bson document from the given pairs of
// field names and values, which can be used as handler command msg.
//
// If any of pair values is nil it's ignored.
func prepareDocument(pairs ...any) (*wirebson.Document, error) {
	l := len(pairs)

	if l%2 != 0 {
		return nil, lazyerrors.Errorf("invalid number of arguments: %d", l)
	}

	docPairs := make([]any, 0, l)

	for i := 0; i < l; i += 2 {
		var err error

		key := pairs[i]
		v := pairs[i+1]

		switch val := v.(type) {
		// json.RawMessage is the non-pointer exception.
		// Other non-pointer types don't need special handling.
		case json.RawMessage:
			v, err = unmarshalSingleJSON(&val)
			if err != nil {
				return nil, err
			}

		case *json.RawMessage:
			if val == nil {
				continue
			}

			v, err = unmarshalSingleJSON(val)
			if err != nil {
				return nil, err
			}
		case *float32:
			if val == nil {
				continue
			}

			v = float64(*val)
		case *bool:
			if val == nil {
				continue
			}

			v = *val
		}

		if v == nil {
			continue
		}

		docPairs = append(docPairs, key, v)
	}

	return wirebson.NewDocument(docPairs...)
}
