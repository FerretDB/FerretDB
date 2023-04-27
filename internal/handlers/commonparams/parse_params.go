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

package commonparams

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/FerretDB/FerretDB/internal/types"
)

// Unmarshal unmarshals a document into a struct.
func Unmarshal(doc *types.Document, value any) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("Unmarshal: value must be a non-nil pointer")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("Unmarshal: value must be a struct pointer")
	}

	// Iterate over the fields of the struct.
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Type().Field(i)

		tag := field.Tag.Get("name")
		if tag == "" {
			tag = field.Name
		}

		// Try to get the value of the field from the document.
		val, err := doc.Get(tag)
		if err != nil {
			return fmt.Errorf("Unmarshal: %s", err)
		}

		// Set the value of the field from the document.
		fv := elem.Field(i)
		if !fv.CanSet() {
			return fmt.Errorf("Unmarshal: field %s is not settable", field.Name)
		}

		if val != nil {
			v := reflect.ValueOf(val)
			if fv.Type() != v.Type() {
				return fmt.Errorf("Unmarshal: field %s type mismatch: got %s, expected %s",
					field.Name, v.Type(), fv.Type())
			}

			fv.Set(v)
		}
	}

	return nil
}
