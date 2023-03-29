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

package conninfo

import (
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// Cursor allows clients to iterate over a result set.
type cursor struct {
	*NewCursorParams

	token *resource.Token
}

// NewCursorParams contains the parameters for creating a new cursor.
type NewCursorParams struct {
	Iter       types.DocumentsIterator
	DB         string
	Collection string
	BatchSize  int32
}

// NewCursor creates a new cursor.
func NewCursor(params *NewCursorParams) *cursor {
	c := &cursor{
		NewCursorParams: params,
		token:           resource.NewToken(),
	}

	resource.Track(c, c.token)

	return c
}

// Close closes the cursor.
func (c *cursor) Close() {
	c.Iter.Close()
	resource.Untrack(c, c.token)
}
