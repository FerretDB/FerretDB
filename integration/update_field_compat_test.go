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

	"go.mongodb.org/mongo-driver/bson"
)

func TestUpdateFieldCompatInc(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Double": {
			update: bson.D{{"$inc", bson.D{{"v", 42.13}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/972",
		},
		"DoubleNegative": {
			update: bson.D{{"$inc", bson.D{{"v", -42.13}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/972",
		},
		"EmptyUpdatePath": {
			update: bson.D{{"$inc", bson.D{{}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/673",
		},
		"DotNotationFieldExist": {
			update: bson.D{{"$inc", bson.D{{"v.foo", int32(1)}}}},
			skip:   "TODO",
		},
		"DotNotationFieldNotExist": {
			update: bson.D{{"$inc", bson.D{{"foo.bar", int32(1)}}}},
			skip:   "TODO",
		},
		"DotNotationArrayFieldExist": {
			update: bson.D{{"$inc", bson.D{{"v.array.0", int32(1)}}}},
			skip:   "TODO",
		},
		"DotNotationArrayFieldNotExist": {
			update: bson.D{{"$inc", bson.D{{"foo.0.baz", int32(1)}}}},
			skip:   "TODO",
		},
	}

	testUpdateCompat(t, testCases)
}
