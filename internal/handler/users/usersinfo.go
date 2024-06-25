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

package users

import (
	"fmt"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// UserFilter creates a filter for finding users.
//
// If `forAllDBs` field is present, a filter to match all users in all databases is returned.
// If number is set, a filter to match all users in given `dbName` is returned.
// Otherwise, it builds filter to match user IDs specified in `doc`.
func UserFilter(doc *bson.Document, dbName string) (*bson.Document, error) {
	usersInfoV := doc.Get(doc.Command())
	userIDs := must.NotFail(bson.NewArray())

	switch usersInfo := usersInfoV.(type) {
	case bson.RawDocument:
		userInfoDoc, err := usersInfo.Decode()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if userInfoDoc.Get("forAllDBs") != nil {
			return must.NotFail(bson.NewDocument()), nil
		}

		userID, err := toUserID(usersInfo, dbName)
		if err != nil {
			return nil, err
		}

		if err = userIDs.Add(userID); err != nil {
			return nil, lazyerrors.Error(err)
		}
	case bson.RawArray:
		usersInfoArr, err := usersInfo.Decode()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		for i := range usersInfoArr.Len() {
			var userID string

			if userID, err = toUserID(usersInfoArr.Get(i), dbName); err != nil {
				return nil, err
			}

			if err = userIDs.Add(userID); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
	case float64, int32, int64:
		return must.NotFail(bson.NewDocument("db", must.NotFail(bson.NewDocument("$eq", dbName)))), nil
	case string:
		userID, err := toUserID(usersInfo, dbName)
		if err != nil {
			return nil, err
		}

		if err = userIDs.Add(userID); err != nil {
			return nil, lazyerrors.Error(err)
		}
	default:
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrBadValue,
			"UserName must be either a string or an object",
			doc.Command(),
		)
	}

	if userIDs.Len() == 0 {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrBadValue,
			"$and/$or/$nor must be a nonempty array",
			doc.Command(),
		)
	}

	return must.NotFail(bson.NewDocument("_id", must.NotFail(bson.NewDocument("$in", userIDs)))), nil
}

// UsersInfo returns users in the format used for usersInfo response from the given batch.
func UsersInfo(batchRaw bson.RawArray, showCredentials bool) (*bson.Array, error) {
	batch, err := batchRaw.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := must.NotFail(bson.NewArray())

	for i := range batch.Len() {
		var user *bson.Document

		if user, err = batch.Get(i).(bson.RawDocument).Decode(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		userInfo := must.NotFail(bson.NewDocument(
			"_id", user.Get("_id"),
			"userId", user.Get("userId"),
			"user", user.Get("user"),
			"db", user.Get("db"),
		))

		var credentials *bson.Document

		if credentials, err = user.Get("credentials").(bson.RawDocument).Decode(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		mechanisms := must.NotFail(bson.NewArray())
		for _, fieldName := range credentials.FieldNames() {
			if err = mechanisms.Add(fieldName); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		if showCredentials {
			if err = userInfo.Add("credentials", credentials); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}

		if err = userInfo.Add("roles", user.Get("roles")); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if err = userInfo.Add("mechanisms", mechanisms); err != nil {
			return nil, lazyerrors.Error(err)
		}

		var userInfoRaw bson.RawDocument

		if userInfoRaw, err = userInfo.Encode(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if err = res.Add(userInfoRaw); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	return res, nil
}

// toUserID returns _id by concatenating db and user field values with a dot.
func toUserID(usersInfoV any, dbName string) (string, error) {
	var db, username string

	switch usersInfo := usersInfoV.(type) {
	case bson.RawDocument:
		usersInfoDoc, err := usersInfo.Decode()
		if err != nil {
			return "", lazyerrors.Error(err)
		}

		userInfo := usersInfoDoc.Get("user")
		if userInfo == nil {
			return "", handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				"UserName must contain a field named: user",
			)
		}

		var ok bool
		if username, ok = userInfo.(string); !ok {
			return "", handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				fmt.Sprintf("UserName must contain a string field named: user. But, has type %T", userInfo),
			)
		}

		dbV := usersInfoDoc.Get("db")
		if dbV == nil {
			return "", handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				"UserName must contain a field named: db",
			)
		}

		if db, ok = dbV.(string); !ok {
			return "", handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				fmt.Sprintf("UserName must contain a string field named: db. But, has type %T", dbV),
			)
		}
	case string:
		username, db = usersInfo, dbName
	default:
		return "", lazyerrors.Errorf("unexpected type %T", usersInfo)
	}

	return db + "." + username, nil
}
