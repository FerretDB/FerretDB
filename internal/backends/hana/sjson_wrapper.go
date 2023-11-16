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

package hana

import (
	"regexp"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
)

const hanaSchemaSymbol = "_$"

func UnmarshalHana(data []byte) (*types.Document, error) {

	re := regexp.MustCompile(`\$([a-z])`)
	replacedData := re.ReplaceAllString(string(data), hanaSchemaSymbol+"$1")

	doc, err := sjson.Unmarshal([]byte(replacedData))

	return doc, err
}

func MarshalHana(doc *types.Document) ([]byte, error) {
	data, err := sjson.Marshal(doc)
	if err != nil {
		return nil, err
	}
	str := string(data)

	re := regexp.MustCompile(`\$([a-z])`)
	replacedStr := re.ReplaceAllString(str, hanaSchemaSymbol+"$1")

	return []byte(replacedStr), nil
}
