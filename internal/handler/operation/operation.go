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

// Package operation provides access to operation registry.
package operation

import (
	"time"

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/resource"
)

// Operation stores information about an operation.
type Operation struct {
	// the order of the fields is weird to reduce size
	Command       *wirebson.Document
	CurrentOpTime time.Time
	token         *resource.Token
	Op            string
	DB            string
	Collection    string
	OpID          int32
	Active        bool
}

// newOperation creates a new operation.
func newOperation(id int32, op string) *Operation {
	o := &Operation{
		token:         resource.NewToken(),
		Op:            op,
		Active:        true,
		CurrentOpTime: time.Now(),
		OpID:          id,
	}

	resource.Track(o, o.token)

	return o
}

// close untracks the operation.
func (o *Operation) close() {
	resource.Untrack(o, o.token)
}
