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
	"fmt"
	"reflect"
	"strings"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// ExtractParams fill passed value structure with parameters from the document.
// If the passed value is not a pointer to the structure it panics.
// Parameters are extracted by the field name or by the `name` tag.
//
// Possible tags:
//   - `opt` - field is optional;
//   - `non-default` - field is unimplemented non-default;
//   - `unimplemented` - field is not implemented yet;
//   - `ignored` - field is ignored.
//
// Collection field processed in a special way. For the commands that require collection name
// it is extracted from the command name.
func ExtractParams(doc *types.Document, command string, value any, l *zap.Logger) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic("value must be a non-nil pointer")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		panic("value must be a struct pointer")
	}

	var i int

	// Iterate over the fields of the struct.
	for ; i < elem.NumField(); i++ {
		field := elem.Type().Field(i)

		tag := field.Tag.Get("name")
		if tag == "" {
			return lazyerrors.Errorf("no tag provided for %s", field.Name)
		}

		var optional, nonDefault, unImplemented, ignored bool

		for _, tt := range strings.Split(tag, ",") {
			switch tt {
			case "opt":
				optional = true
			case "non-default":
				nonDefault = true
			case "unimplemented":
				unImplemented = true
			case "ignored":
				ignored = true
			default:
				tag = tt

				if tt == "collection" {
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

			return lazyerrors.Error(err)
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
			return lazyerrors.Errorf("field %s is not settable", field.Name)
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
			return lazyerrors.Errorf("field %s type %s is not supported", field.Name, fv.Type())
		}

		if settable != nil {
			v := reflect.ValueOf(settable)

			if v.Type() != fv.Type() {
				if tag == command {
					return commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrInvalidNamespace,
						fmt.Sprintf("collection name has invalid type %s", AliasFromType(settable)),
						command,
					)
				}

				return lazyerrors.Errorf(
					"field %s type mismatch: got %s, expected %s",
					field.Name, v.Type(), fv.Type(),
				)
			}

			fv.Set(v)

			continue
		}

		if val != nil {
			v := reflect.ValueOf(val)
			if fv.Type() != v.Type() {
				return lazyerrors.Errorf(
					"field %s type mismatch: got %s, expected %s",
					field.Name, v.Type(), fv.Type(),
				)
			}

			fv.Set(v)
		}
	}

	if i != doc.Len() {
		// some unprocessed fields left
	}

	return nil
}
