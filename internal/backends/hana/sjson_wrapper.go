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

	"github.com/FerretDB/FerretDB/internal/handler/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
)

const hanaSchemaSymbol = "_$"

func hanaJSONToStr(data []byte) []byte {
	str := string(data)

	re := regexp.MustCompile(regexp.QuoteMeta(hanaSchemaSymbol) + `([a-z])`)
	replacedStr := re.ReplaceAllString(str, "$$$1")

	return []byte(replacedStr)
}

func strToHanaJSON(data []byte) []byte {
	str := string(data)

	re := regexp.MustCompile(`\$([a-z])`)
	replacedStr := re.ReplaceAllString(str, hanaSchemaSymbol+"$$1")

	return []byte(replacedStr)
}

func unmarshalHana(data []byte) (*types.Document, error) {
	replacedData := hanaJSONToStr(data)
	doc, err := sjson.Unmarshal([]byte(replacedData))

	return doc, err
}

func marshalHana(doc *types.Document) ([]byte, error) {
	data, err := sjson.Marshal(doc)
	if err != nil {
		return nil, err
	}

	return strToHanaJSON(data), nil
}
