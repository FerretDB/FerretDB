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

	"github.com/FerretDB/wire/wirebson"
	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/v2/internal/handler/session"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// getRequiredParamAny returns doc's first value for the given key
// or protocol error for missing key.
func getRequiredParamAny(doc wirebson.AnyDocument, key string) (any, error) {
	d, err := doc.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	v := d.Get(key)
	if v == nil {
		msg := fmt.Sprintf("required parameter %q is missing", key)
		return nil, lazyerrors.Error(mongoerrors.NewWithArgument(mongoerrors.ErrBadValue, msg, key))
	}

	return v, nil
}

// getRequiredParam returns doc's first value for the given key
// or protocol error for missing key or invalid value type.
func getRequiredParam[T wirebson.ScalarType](doc wirebson.AnyDocument, key string) (T, error) {
	var zero T

	v, err := getRequiredParamAny(doc, key)
	if err != nil {
		return zero, lazyerrors.Error(err)
	}

	res, ok := v.(T)
	if !ok {
		msg := fmt.Sprintf("required parameter %q has type %T (expected %T)", key, v, zero)
		return zero, lazyerrors.Error(mongoerrors.NewWithArgument(mongoerrors.ErrBadValue, msg, key))
	}

	return res, nil
}

// getOptionalParamAny returns doc's first value for the given key.
// If the value is missing, it returns a default value.
func getOptionalParamAny(doc wirebson.AnyDocument, key string, defaultValue any) (any, error) {
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
func getOptionalParam[T wirebson.ScalarType](doc wirebson.AnyDocument, key string, defaultValue T) (T, error) {
	var zero T

	v, err := getOptionalParamAny(doc, key, defaultValue)
	if err != nil {
		return zero, lazyerrors.Error(err)
	}

	res, ok := v.(T)
	if !ok {
		msg := fmt.Sprintf("required parameter %q has type %T (expected %T)", key, v, zero)
		return zero, lazyerrors.Error(mongoerrors.NewWithArgument(mongoerrors.ErrBadValue, msg, key))
	}

	return res, nil
}

// getBoolParam returns bool value of v.
// Non-zero double, long, and int values return true.
// Zero values for those types, as well as nulls and missing fields, return false.
// Other types return a protocol error.
func getBoolParam(key string, v any) (bool, error) {
	switch v := v.(type) {
	case float64:
		return v != 0, nil
	case bool:
		return v, nil
	case wirebson.NullType:
		return false, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	default:
		msg := fmt.Sprintf(
			`BSON field '%s' is the wrong type '%s', expected types '[bool, long, int, decimal, double]'`,
			key,
			aliasFromType(v),
		)

		return false, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, key)
	}
}

// getSessionIDsParam returns session UUIDs from the document.
// The document has the format `{<key>: [{id: <uuid>}, ...]}` and
// a protocol error is returned for invalid format or value.
func getSessionIDsParam(doc wirebson.AnyDocument, key string) ([]uuid.UUID, error) {
	d, err := doc.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	v := d.Get(key)

	sessionsArray, ok := v.(wirebson.AnyArray)
	if !ok {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrTypeMismatch,
			fmt.Sprintf("BSON field '%[1]s.%[1]s' is the wrong type '%[2]T', expected type 'array'", key, v),
			key,
		)
	}

	sessions, err := sessionsArray.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	ids := make([]uuid.UUID, sessions.Len())

	for i, v := range sessions.All() {
		var sessionDoc wirebson.AnyDocument

		if sessionDoc, ok = v.(wirebson.AnyDocument); !ok {
			m := fmt.Sprintf(
				"BSON field '%[1]s.%[1]sFromClient.%[2]d' is the wrong type '%[3]T', expected type 'object'",
				key,
				i,
				v,
			)

			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, m, key)
		}

		var session *wirebson.Document

		if session, err = sessionDoc.Decode(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if v = session.Get("id"); v == nil {
			m := fmt.Sprintf("BSON field '%[1]s.%[1]sFromClient.id' is missing but a required field", key)
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrLocation40414, m, key)
		}

		var id wirebson.Binary

		if id, ok = v.(wirebson.Binary); !ok {
			m := fmt.Sprintf(
				"BSON field '%[1]s.%[1]sFromClient.id' is the wrong type '%[2]T', expected type 'binData'",
				key,
				v,
			)

			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, m, key)
		}

		if id.Subtype != wirebson.BinaryUUID {
			m := fmt.Sprintf(
				"BSON field '%[1]s.%[1]sFromClient.id' is the wrong binData type '%[2]s', expected type 'UUID'",
				key,
				id.Subtype.String(),
			)

			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, m, key)
		}

		var sessionID uuid.UUID

		if sessionID, err = uuid.FromBytes(id.B); err != nil {
			m := "uuid must be a 16-byte binary field with UUID (4) subtype"
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrInvalidUUID, m, key)
		}

		ids[i] = sessionID
	}

	return ids, nil
}

// getSessionUsersParam returns User IDs from given db and user pairs.
// The `v` has the format `[{db:<dbname>, user:<username>}, ...]` and
// a protocol error is returned for invalid format or value.
func getSessionUsersParam(v any, command, field string) ([]session.UserID, error) {
	usersV, ok := v.(wirebson.AnyArray)
	if !ok {
		msg := fmt.Sprintf("BSON field '%s' is the wrong type '%T', expected type 'array'", field, v)
		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, command)
	}

	usersArr, err := usersV.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	userIDs := make([]session.UserID, usersArr.Len())

	for i, v := range usersArr.All() {
		var userDoc wirebson.AnyDocument

		if userDoc, ok = v.(wirebson.AnyDocument); !ok {
			msg := fmt.Sprintf("BSON field '%s.%d' is the wrong type '%T', expected type 'object'", field, i, v)
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, command)
		}

		var user *wirebson.Document

		if user, err = userDoc.Decode(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if v = user.Get("db"); v == nil {
			msg := fmt.Sprintf("BSON field '%s.db' is missing but a required field", field)
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrLocation40414, msg, command)
		}

		var dbName, username string

		if dbName, ok = v.(string); !ok {
			msg := fmt.Sprintf("BSON field '%s.db' is the wrong type '%T', expected type 'string'", field, v)
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, command)
		}

		if v = user.Get("user"); v == nil {
			msg := fmt.Sprintf("BSON field '%s.user' is missing but a required field", field)
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrLocation40414, msg, command)
		}

		if username, ok = v.(string); !ok {
			msg := fmt.Sprintf("BSON field '%s.user' is the wrong type '%T', expected type 'string'", field, v)
			return nil, mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, command)
		}

		userIDs[i] = session.GetUIDFromUsername(dbName, username)
	}

	return userIDs, nil
}
