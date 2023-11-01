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
)

func TestDiffDebugError(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct {
		input string

		err *mongo.CommandError
		res bson.D
	}{
		"Empty": {
			input: "",
			err: &mongo.CommandError{
				Code:    1,
				Name:    "InternalError",
				Message: "command failed",
			},
		},
		"Ok": {
			input: "ok",
			res:   bson.D{{"ok", float64(1)}},
		},
		"CodeZero": {
			input: "0",
			err: &mongo.CommandError{
				Code:    1,
				Name:    "InternalError",
				Message: "Unset",
			},
		},
		"CodeOne": {
			input: "1",
			err: &mongo.CommandError{
				Code:    1,
				Name:    "InternalError",
				Message: "InternalError",
			},
		},
		"CodeNotLabeled": {
			input: "33333",
			err: &mongo.CommandError{
				Code:    1,
				Name:    "InternalError",
				Message: "ErrorCode(33333)",
			},
		},
		// TODO https://github.com/FerretDB/FerretDB/issues/3672
		//"Panic": {
		//	input: "panic",
		//	ferretErr: &mongo.CommandError{
		//		Code:    0,
		//		Message: "connection(127.0.0.1:27017[-22]) socket was unexpectedly closed: EOF",
		//		Labels:  []string{"NetworkError"},
		//	},
		//},
		"String": {
			input: "foo",
			err: &mongo.CommandError{
				Code:    1,
				Name:    "InternalError",
				Message: "foo",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var res bson.D
			err := collection.Database().RunCommand(ctx, bson.D{{"debugError", tc.input}}).Decode(&res)

			if setup.IsMongoDB(t) {
				expected := mongo.CommandError{
					Code:    59,
					Message: "no such command: 'debugError'",
					Name:    "CommandNotFound",
				}
				AssertEqualCommandError(t, expected, err)

				return
			}

			if tc.err != nil {
				AssertEqualCommandError(t, *tc.err, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.res, res)
		})
	}
}
