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

package integration

import (
	"testing"

	"github.com/FerretDB/FerretDB/integration/setup"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestGetDistinctCommandInvalidQuery(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	db := collection.Database()

	_, err := db.RunCommand(
		ctx, bson.D{{"distinct", collection.Name()}, {"key", "a"}, {"query", 1}},
	).DecodeBytes()

	AssertEqualError(t, mongo.CommandError{
		Code:    14,
		Name:    "TypeMismatch",
		Message: `BSON field 'query' is the wrong type 'int', expected type 'object'`},
		err,
	)
}
