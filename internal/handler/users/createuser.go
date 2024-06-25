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
	"strings"

	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/password"
)

// UserDocument returns a document to insert into the user collection.
func UserDocument(mechanisms bson.AnyArray, dbName, username, userPassword string) (*bson.Document, error) {
	if dbName == "$external" {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrNotImplemented,
			"createUser for $external database is not implemented",
			"createUser",
		)
	}

	if username == "" {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrBadValue,
			"User document needs 'user' field to be non-empty",
			"createUser",
		)
	}

	if userPassword == "" {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrSetEmptyPassword,
			"Password cannot be empty",
			"createUser",
		)
	}

	credentials, err := MakeCredentials("createUser", username, userPassword, mechanisms)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return bson.NewDocument(
		"_id", dbName+"."+username,
		"credentials", credentials,
		"user", username,
		"db", dbName,
		"roles", bson.MakeArray(0),
		"userId", bson.Binary{
			Subtype: bson.BinaryUUID,
			B:       must.NotFail(uuid.New().MarshalBinary()),
		},
	)
}

// MakeCredentials returns a document with credentials for the chosen mechanisms.
func MakeCredentials(command, username, userPassword string, mechanismsAny bson.AnyArray) (*bson.Document, error) {
	mechanisms, err := mechanismsAny.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if mechanisms.Len() == 0 {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrBadValue,
			"mechanisms field must not be empty",
			command,
		)
	}

	credentials := bson.MakeDocument(0)

	for i := range mechanisms.Len() {
		mechanism := mechanisms.Get(i)
		switch mechanism {
		case "SCRAM-SHA-1":
			var hash *bson.Document

			if hash, err = password.SCRAMSHA1VariationHash(username, userPassword); err != nil {
				return nil, err
			}

			if err = credentials.Add("SCRAM-SHA-1", hash); err != nil {
				return nil, lazyerrors.Error(err)
			}
		case "SCRAM-SHA-256":
			var hash *bson.Document

			if hash, err = password.SCRAMSHA256Hash(userPassword); err != nil {
				if strings.Contains(err.Error(), "prohibited character") {
					return nil, handlererrors.NewCommandErrorMsg(
						handlererrors.ErrStringProhibited,
						"Error preflighting normalization: U_STRINGPREP_PROHIBITED_ERROR",
					)
				}

				return nil, err
			}

			if err = credentials.Add("SCRAM-SHA-256", hash); err != nil {
				return nil, lazyerrors.Error(err)
			}
		default:
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrBadValue,
				fmt.Sprintf("Unknown auth mechanism '%s'", mechanism),
				"createUser",
			)
		}
	}

	return credentials, nil
}
