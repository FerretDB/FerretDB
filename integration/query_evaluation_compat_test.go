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
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestQueryEvaluationCompatRegexErrors(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/908")

	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"MissingClosingParen": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "g(-z]+ng  wrong regex"}}}}},
			resultType: emptyResult,
		},
		"MissingClosingBracket": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "g[-z+ng  wrong regex"}}}}},
			resultType: emptyResult,
		},
		"InvalidEscape": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "\\uZ"}}}}},
			resultType: emptyResult,
		},
		"NamedCapture": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "(?P<name)"}}}}},
			resultType: emptyResult,
		},
		"UnexpectedParen": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: ")"}}}}},
			resultType: emptyResult,
		},
		"TrailingBackslash": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `abc\`}}}}},
			resultType: emptyResult,
		},
		"InvalidRepetition": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `a**`}}}}},
			resultType: emptyResult,
		},
		"MissingRepetitionArgumentStar": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `*`}}}}},
			resultType: emptyResult,
		},
		"MissingRepetitionArgumentPlus": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `+`}}}}},
			resultType: emptyResult,
		},
		"MissingRepetitionArgumentQuestion": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `?`}}}}},
			resultType: emptyResult,
		},
		"InvalidClassRange": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `[z-a]`}}}}},
			resultType: emptyResult,
		},
		"InvalidNestedRepetitionOperatorStar": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `a**`}}}}},
			resultType: emptyResult,
		},
		"InvalidPerlOp": {
			filter:     bson.D{{"v", bson.D{{"$regex", `(?z)`}}}},
			resultType: emptyResult,
		},
		"InvalidRepeatSize": {
			filter:     bson.D{{"v", bson.D{{"$regex", `(aa){3,10001}`}}}},
			resultType: emptyResult,
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryEvaluationCompatRegex(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"RegexNoSuchField": {
			filter:     bson.D{{"no-such-field", bson.D{{"$regex", primitive.Regex{Pattern: "foo"}}}}},
			resultType: emptyResult,
		},
		"RegexNoSuchFieldString": {
			filter:     bson.D{{"no-such-field", bson.D{{"$regex", "foo"}}}},
			resultType: emptyResult,
		},
		"RegexBadOption": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "foo", Options: "123"}}}}},
			resultType: emptyResult,
		},
	}

	testQueryCompat(t, testCases)
}
