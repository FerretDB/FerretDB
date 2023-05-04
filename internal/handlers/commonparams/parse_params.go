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
	"strings"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
)

// Unmarshal unmarshals a document into a struct.
func Unmarshal(doc *types.Document, command string, value any, l *zap.Logger) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("unmarshal: value must be a non-nil pointer")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("unmarshal: value must be a struct pointer")
	}

	// Iterate over the fields of the struct.
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Type().Field(i)

		tag := field.Tag.Get("name")
		if tag == "" {
			tag = field.Name
		}

		var optional, nonDefault, unImplemented, ignored bool

		for _, i := range strings.Split(tag, ",") {
			switch i {
			case "opt":
				optional = true
			case "non-default":
				nonDefault = true
			case "unimplemented":
				unImplemented = true
			case "ignored":
				ignored = true
			default:
				tag = i

				if i == "collection" {
					tag = command
				}
			}
		}

		// Try to get the value of the field from the document.
		val, err := doc.Get(tag)
		if err != nil {
			if optional || nonDefault || unImplemented || ignored {
				continue
			}

			return fmt.Errorf("unmarshal: %s", err)
		}

		if ignored {
			l.Debug(
				"ignoring field",
				zap.String("command", doc.Command()), zap.String("field", tag), zap.Any("value", val),
			)

			continue
		}

		if unImplemented {
			msg := fmt.Sprintf(
				"%s: support for field %q with value %v is not implemented yet",
				doc.Command(), tag, val,
			)

			return commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNotImplemented, msg, tag)
		}

		if nonDefault {
			v, ok := val.(bool)

			if ok && !v {
				msg := fmt.Sprintf(
					"%s: support for field %q with non-default value %v is not implemented yet",
					doc.Command(), tag, val,
				)

				return commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNotImplemented, msg, tag)
			}
		}

		// Set the value of the field from the document.
		fv := elem.Field(i)
		if !fv.CanSet() {
			return fmt.Errorf("unmarshal: field %s is not settable", field.Name)
		}

		var settable any

		switch fv.Kind() { //nolint: exhaustive // TODO: add more types support
		case reflect.Int32, reflect.Int64, reflect.Float64:
			if tag == "skip" || tag == "limit" {
				settable, err = GetWholeParamStrict(command, tag, val)
				if err != nil {
					return err
				}

				break
			}

			settable, err = GetWholeNumberParam(val)
			if err != nil {
				return err
			}
		case reflect.String, reflect.Bool, reflect.Struct, reflect.Pointer:
			settable = val
		default:
			return fmt.Errorf("unmarshal: field %s type %s is not supported", field.Name, fv.Type())
		}

		if settable != nil {
			v := reflect.ValueOf(settable)
			fv.Set(v)

			continue
		}

		if val != nil {
			v := reflect.ValueOf(val)
			if fv.Type() != v.Type() {
				return fmt.Errorf(
					"unmarshal: field %s type mismatch: got %s, expected %s",
					field.Name, v.Type(), fv.Type(),
				)
			}

			fv.Set(v)
		}
	}

	return nil
}
