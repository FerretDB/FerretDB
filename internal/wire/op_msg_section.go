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

package wire

import (
	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// opMsgSection is one or more sections contained in an OpMsg.
type opMsgSection struct {
	// The order of fields is weird to make the struct smaller due to alignment.
	// The wire order is: kind, identifier, documents.

	identifier string
	documents  []bson.RawDocument
	kind       byte
}

// checkSections checks given sections.
func checkSections(sections []opMsgSection) error {
	if len(sections) == 0 {
		return lazyerrors.New("no sections")
	}

	var kind0Found bool

	for _, s := range sections {
		switch s.kind {
		case 0:
			if kind0Found {
				return lazyerrors.New("multiple kind 0 sections")
			}
			kind0Found = true

			if s.identifier != "" {
				return lazyerrors.New("kind 0 section has identifier")
			}

			if len(s.documents) != 1 {
				return lazyerrors.Errorf("kind 0 section has %d documents", len(s.documents))
			}

		case 1:
			if s.identifier == "" {
				return lazyerrors.New("kind 1 section has no identifier")
			}

		default:
			return lazyerrors.Errorf("unknown kind %d", s.kind)
		}
	}

	return nil
}
