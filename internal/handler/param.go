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

package handler

import (
	"fmt"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// getRequiredParamAny returns doc's first value for the given key
// or protocol error for missing key.
func getRequiredParamAny(doc bson.AnyDocument, key string) (any, error) {
	d, err := doc.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	v := d.Get(key)
	if v == nil {
		msg := fmt.Sprintf("required parameter %q is missing", key)
		return nil, lazyerrors.Error(handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrBadValue, msg, key))
	}

	return v, nil
}

// getRequiredParam returns doc's first value for the given key
// or protocol error for missing key or invalid value type.
func getRequiredParam[T bson.ScalarType](doc bson.AnyDocument, key string) (T, error) {
	var zero T

	v, err := getRequiredParamAny(doc, key)
	if err != nil {
		return zero, lazyerrors.Error(err)
	}

	res, ok := v.(T)
	if !ok {
		msg := fmt.Sprintf("required parameter %q has type %T (expected %T)", key, v, zero)
		return zero, lazyerrors.Error(handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrBadValue, msg, key))
	}

	return res, nil
}

// getOptionalParamAny returns doc's first value for the given key.
// If the value is missing, it returns a default value.
func getOptionalParamAny(doc bson.AnyDocument, key string, defaultValue any) (any, error) {
	d, err := doc.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	v := d.Get(key)
	if v == nil {
		return defaultValue, nil
	}

	return v, nil
}

// getOptionalParam returns doc's first value for the given key
// or protocol error for invalid value type.
// If the value is missing, it returns a default value.
func getOptionalParam[T bson.ScalarType](doc bson.AnyDocument, key string, defaultValue T) (T, error) {
	var zero T

	v, err := getOptionalParamAny(doc, key, defaultValue)
	if err != nil {
		return zero, lazyerrors.Error(err)
	}

	res, ok := v.(T)
	if !ok {
		msg := fmt.Sprintf("required parameter %q has type %T (expected %T)", key, v, zero)
		return zero, lazyerrors.Error(handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrBadValue, msg, key))
	}

	return res, nil
}
