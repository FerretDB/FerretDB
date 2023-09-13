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

package hanadb

import (
	"bytes"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
)

func marshal(doc *types.Document) ([]byte, error) {
	bdoc, err := sjson.Marshal(doc)

	if err != nil {
		return nil, err
	}

	return bytes.Replace(bdoc, []byte("$"), []byte("%%DollarSign%%"), -1), nil
}

func unmarshal(data []byte) (*types.Document, error) {
	return sjson.Unmarshal(bytes.Replace(data, []byte("%%DollarSign%%"), []byte("$"), -1))
}
