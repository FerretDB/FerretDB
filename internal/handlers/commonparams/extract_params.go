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
	"slices"
	"strings"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// ExtractParams fill passed value structure with parameters from the document.
// If the passed value is not a pointer to the structure it panics.
// Parameters are extracted by the field name or by the `ferretdb` tag.
//
// Possible tags:
//   - `opt` - field is optional, the field value would not be set if it's not present in the document;
//   - `unimplemented-non-default` - error would be returned if non-default value for the field is provided;
//   - `unimplemented` - error would be returned if the value is present in the document;
//   - `ignored` - field is ignored and would not be set, but it would be logged;
//   - `positiveNumber` - provided value must be of types [int, long, double] and greater than 0,
//     double values would be rounded to long;
//   - `wholePositiveNumber` - provided value must be of types [int, long] and greater than 0;
//   - `numericBool` - provided value must be of types [bool, int, long, double] and would be converted to bool;
//   - `zeroOrOneAsBool` - provided value must be of types [int, long, double] with possible values `0` or `1`.
//   - `collection` - Collection field value holds the name of the collection and must be of type string. An error is
//     returned if `collection` tag is not set.
//
// It returns command errors with the following codes:
//   - `ErrFailedToParse` when provided field is not present in passed structure;
//   - `ErrFailedToParse` when provided field must be 0 or 1, but it is not;
//   - `ErrNotImplemented` when support for provided field is not implemented yet;
//   - `ErrNotImplemented`when support for non-default field value is not implemented yet;
//   - `ErrValueNegative` - field of numeric type is negative;
//   - `ErrTypeMismatch` - field is type is not matched with the type of the value;
//   - `ErrBadValue` when field is not a number;
//   - `ErrBadValue` when field is not of integer type;
//   - `ErrBadValue` when field is out of integer range;
//   - `ErrBadValue` when field has negative value;
//   - `ErrInvalidNamespace` - collection name has invalid type.
func ExtractParams(doc *types.Document, command string, value any, l *zap.Logger) error {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic("value must be a non-nil pointer")
	}

	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		panic("value must be a struct pointer")
	}

	iter := doc.Iterator()
	defer iter.Close()

	for {
		key, val, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		lookup := key

		// If the key is the same as the command name, then it is a collection name.
		// Depending on the driver, the key may be camel case or lower case for a collection name,
		// hence use camel cased `command` for lookup of camel cased field tag.
		if strings.ToLower(key) == strings.ToLower(command) { //nolint:staticcheck // for clarity
			lookup = command
		}

		fieldIndex, options, err := lookupFieldTag(lookup, &elem)
		if err != nil {
			panic(err)
		}

		if fieldIndex == nil {
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrFailedToParse,
				fmt.Sprintf("%s: unknown field %q", command, key),
				command,
			)
		}

		if options.ignored {
			l.Debug(
				"ignoring field",
				zap.String("command", doc.Command()), zap.String("field", key), zap.Any("value", val),
			)

			continue
		}

		if options.unimplemented {
			msg := fmt.Sprintf(
				"%s: support for field %q with value %v is not implemented yet",
				doc.Command(), key, val,
			)

			return commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNotImplemented, msg, key)
		}

		if options.nonDefault {
			v, ok := val.(bool)

			if ok && v {
				msg := fmt.Sprintf(
					"%s: support for field %q with non-default value %v is not implemented yet",
					doc.Command(), key, val,
				)

				return commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNotImplemented, msg, key)
			}
		}

		err = setStructField(&elem, options, *fieldIndex, command, key, val, l)
		if err != nil {
			return err
		}
	}

	err := checkAllRequiredFieldsPopulated(&elem, command, doc.Keys())
	if err != nil {
		return err
	}

	return nil
}

// tagOptions contains options for the structure field tag.
type tagOptions struct {
	optional            bool
	nonDefault          bool
	unimplemented       bool
	ignored             bool
	positiveNumber      bool
	wholePositiveNumber bool
	numericBool         bool
	zeroOrOneAsBool     bool
	collection          bool
}

// lookupFieldTag looks for the tag and returns its options.
func lookupFieldTag(key string, value *reflect.Value) (*int, *tagOptions, error) {
	var to *tagOptions
	var i int
	var found bool

	for ; i < value.NumField(); i++ {
		field := value.Type().Field(i)

		tag := field.Tag.Get("ferretdb")

		if tag == "" {
			return nil, nil, lazyerrors.Errorf("no tag provided for %s", field.Name)
		}

		optionsList := strings.Split(tag, ",")

		if optionsList[0] != key {
			continue
		}

		to = tagOptionsFromList(optionsList[1:])

		found = true

		break
	}

	if !found {
		return nil, nil, nil
	}

	return &i, to, nil
}

func tagOptionsFromList(optionsList []string) *tagOptions {
	var to tagOptions

	for _, tt := range optionsList {
		switch tt {
		case "opt":
			to.optional = true
		case "unimplemented-non-default":
			to.nonDefault = true
		case "unimplemented":
			to.unimplemented = true
		case "ignored":
			to.ignored = true
		case "numericBool":
			to.numericBool = true
		case "positiveNumber":
			to.positiveNumber = true
		case "wholePositiveNumber":
			to.wholePositiveNumber = true
		case "zeroOrOneAsBool":
			to.zeroOrOneAsBool = true
		case "collection":
			to.collection = true
		default:
			panic(fmt.Sprintf("unknown tag option %s", tt))
		}
	}

	return &to
}

// setStructField sets the value of the document field to the structure field.
func setStructField(elem *reflect.Value, o *tagOptions, i int, command, key string, val any, l *zap.Logger) error {
	var err error

	field := elem.Type().Field(i)

	// Set the value of the field from the document.
	fv := elem.Field(i)
	if !fv.CanSet() {
		panic(fmt.Sprintf("field %s is not settable", field.Name))
	}

	var settable any

	switch fv.Kind() { //nolint: exhaustive // all other types are not supported
	case reflect.Int32, reflect.Int64, reflect.Float64:
		if o.positiveNumber {
			settable, err = GetValidatedNumberParamWithMinValue(command, key, val, 0)
			if err != nil {
				return err
			}

			break
		}

		if o.wholePositiveNumber {
			settable, err = getOptionalPositiveNumber(key, val)
			if err != nil {
				return err
			}

			break
		}

		settable, err = GetWholeNumberParam(val)
		if err != nil {
			return err
		}
	case reflect.String, reflect.Struct, reflect.Pointer, reflect.Interface:
		settable = val
	case reflect.Bool:
		if o.numericBool {
			settable, err = GetBoolOptionalParam(key, val)
			if err != nil {
				return err
			}

			break
		}

		if o.zeroOrOneAsBool {
			var numeric int64

			numeric, err = GetWholeNumberParam(val)
			if err != nil || numeric < 0 || numeric > 1 {
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrFailedToParse,
					fmt.Sprintf("The '%s.%s' field must be 0 or 1. Got %v", command, key, types.FormatAnyValue(val)),
					command,
				)
			}

			settable = numeric == 1

			break
		}

		settable = val
	case reflect.Slice:
		array, ok := val.(*types.Array)
		if !ok {
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				fmt.Sprintf(
					`BSON field '%s.%s' is the wrong type '%s', expected type '%s'`,
					command, key, AliasFromType(val), AliasFromType(fv.Interface()),
				),
				command,
			)
		}

		iter := array.Iterator()
		defer iter.Close()

		arrayToSet := reflect.MakeSlice(fv.Type(), array.Len(), array.Len())

		for {
			i, arrayDoc, err := iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			if err != nil {
				return lazyerrors.Error(err)
			}

			doc, ok := arrayDoc.(*types.Document)
			if !ok {
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrTypeMismatch,
					fmt.Sprintf(
						`BSON field '%s.%s' is the wrong type '%s', expected type '%s'`,
						command, key, AliasFromType(val), AliasFromType(fv.Interface()),
					),
					command,
				)
			}

			params := reflect.New(fv.Type().Elem())

			err = ExtractParams(doc, command+"."+key, params.Interface(), l)
			if err != nil {
				return err
			}

			arrayToSet.Index(i).Set(params.Elem())
		}

		settable = arrayToSet.Interface()
	default:
		panic(fmt.Sprintf("field %s type %s is not supported", field.Name, fv.Type()))
	}

	if settable != nil {
		v := reflect.ValueOf(settable)

		if key == command && !o.collection {
			return lazyerrors.New("collection field contains value that is not a collection name")
		}

		if v.Type() != fv.Type() {
			if key == command {
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrInvalidNamespace,
					fmt.Sprintf("collection name has invalid type %s", AliasFromType(settable)),
					command,
				)
			}

			if fv.Kind() == reflect.Interface {
				fv.Set(reflect.ValueOf(v.Interface()))
				return nil
			}

			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrTypeMismatch,
				fmt.Sprintf(
					`BSON field '%s.%s' is the wrong type '%s', expected type '%s'`,
					command, key, AliasFromType(val), AliasFromType(fv.Interface()),
				),
				command,
			)
		}

		fv.Set(v)
	}

	return nil
}

// checkAllRequiredFieldsPopulated checks that all required fields are populated.
func checkAllRequiredFieldsPopulated(v *reflect.Value, command string, keys []string) error {
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)

		tag := field.Tag.Get("ferretdb")

		optionsList := strings.Split(tag, ",")

		if len(optionsList) == 0 {
			return lazyerrors.Errorf("no tag provided for %s", field.Name)
		}

		to := tagOptionsFromList(optionsList[1:])
		if to.ignored || to.optional || to.unimplemented || to.nonDefault {
			continue
		}

		if to.collection {
			if v.Field(i).IsZero() {
				return commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace,
					fmt.Sprintf("Invalid namespace specified '%s.'", v.FieldByName("DB")),
					command,
				)
			}
		}

		key := optionsList[0]

		// Fields with "-" are ignored when parsing parameters.
		if key == "-" {
			continue
		}

		// Depending on the driver, the key may be camel case or lower case for a collection name.
		if !slices.Contains(keys, key) && !slices.Contains(keys, strings.ToLower(key)) {
			return commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrMissingField,
				fmt.Sprintf("BSON field '%s.%s' is missing but a required field", command, key),
				command,
			)
		}
	}

	return nil
}
