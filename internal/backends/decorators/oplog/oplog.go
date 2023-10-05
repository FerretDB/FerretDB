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

// Package oplog provides decorators that add OpLog functionality to the backend.
package oplog

import (
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// document represents a single OpLog collection record.
type document struct {
	o  *types.Document
	ns string
	op string // i, d
}

// marshal returns the BSON document representation with a given timestamp.
func (d *document) marshal(t time.Time) (*types.Document, error) {
	res, err := types.NewDocument(
		"_id", types.NewObjectID(), // TODO
		"ts", types.NextTimestamp(t),
		"ns", d.ns,
		"op", d.op,
		"o", d.o,
		"t", int64(1),
		"v", int64(2),
		"wall", t,
	)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = res.ValidateData(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}
