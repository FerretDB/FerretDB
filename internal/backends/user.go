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

package backends

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/password"
)

// CreateUser stores a new user in the given database and Backend.
func CreateUser(ctx context.Context, b Backend, mechanisms *types.Array, dbName, username, password string) error {
	credentials, err := MakeCredentials(mechanisms, username, password)
	if err != nil {
		return err
	}

	id := uuid.New()
	saved := must.NotFail(types.NewDocument(
		"_id", dbName+"."+username,
		"credentials", credentials,
		"user", username,
		"db", dbName,
		"roles", types.MakeArray(0),
		"userId", types.Binary{Subtype: types.BinaryUUID, B: must.NotFail(id.MarshalBinary())},
	))

	adminDB, err := b.Database("admin")
	if err != nil {
		return err
	}

	users, err := adminDB.Collection("system.users")
	if err != nil {
		return err
	}

	_, err = users.InsertAll(ctx, &InsertAllParams{
		Docs: []*types.Document{saved},
	})

	return err
}

// MakeCredentials creates a document with credentials for the chosen mechanisms.
// The mechanisms array must be validated by the caller.
func MakeCredentials(mechanisms *types.Array, username, userPassword string) (*types.Document, error) {
	credentials := types.MakeDocument(0)

	if mechanisms == nil {
		mechanisms = must.NotFail(types.NewArray("SCRAM-SHA-1", "SCRAM-SHA-256"))
	}

	iter := mechanisms.Iterator()
	defer iter.Close()

	for {
		var v any
		_, v, err := iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		var hash *types.Document

		switch v {
		case "PLAIN":
			credentials.Set("PLAIN", must.NotFail(password.PlainHash(userPassword)))
		case "SCRAM-SHA-1":
			hash, err = password.SCRAMSHA1Hash(username, userPassword)
			if err != nil {
				return nil, err
			}

			credentials.Set("SCRAM-SHA-1", hash)
		case "SCRAM-SHA-256":
			hash, err = password.SCRAMSHA256Hash(userPassword)
			if err != nil {
				return nil, err
			}

			credentials.Set("SCRAM-SHA-256", hash)
		default:
			panic("unknown mechanism")
		}
	}

	return credentials, nil
}
