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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// TestInsertInvalid tries to insert an invalid document.
func TestInsertInvalid(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1248")

	t.Parallel()

	for name, tc := range map[string]struct {
		insert      bson.D
		expectedErr *mongo.CommandError
	}{
		"KeyIsNotUTF8": {
			insert: bson.D{
				//  the key is out of range for UTF-8
				{"\xF4\x90\x80\x80", int32(12)},
			},
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Invalid document, reason: invalid key: "\xf4\x90\x80\x80" (not a valid UTF-8 string).`,
			},
		},
		"KeyContains$": {
			insert: bson.D{
				{"v$", int32(12)},
			},
			expectedErr: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `Invalid document, reason: invalid key: "v$" (key mustn't contain $).`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup.Setup(t, shareddata.Int32s)

			res, err := collection.InsertOne(ctx, tc.insert)
			require.Nil(t, res)
			require.NotNil(t, err)
			AssertEqualError(t, *tc.expectedErr, err)
		})
	}
}
