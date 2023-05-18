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

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestGetDistinctCommandInvalidQuery(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		command    bson.D
		err        *mongo.CommandError
		altMessage string
	}{
		"Int": {
			command: bson.D{{"query", 1}},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: `BSON field 'distinct.query' is the wrong type 'int', expected type 'object'`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, collection := setup.Setup(t, shareddata.Doubles)

			db := collection.Database()

			command := bson.D{{"distinct", collection.Name()}}
			command = append(command, bson.E{"key", "a"})

			command = append(command, tc.command...)
			_, err := db.RunCommand(ctx, command).DecodeBytes()

			assert.EqualError(t, err, tc.err.Error())
		})
	}
}
