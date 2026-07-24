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

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestNewCollation(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		doc          *types.Document
		nilCollation bool
		errMsg       string
	}{
		"Nil": {
			doc:          nil,
			nilCollation: true,
		},
		"Simple": {
			doc:          must.NotFail(types.NewDocument("locale", "simple")),
			nilCollation: true,
		},
		"Locale": {
			doc: must.NotFail(types.NewDocument("locale", "en")),
		},
		"LocaleStrength": {
			doc: must.NotFail(types.NewDocument("locale", "en", "strength", int32(2))),
		},
		"NumericOrdering": {
			doc: must.NotFail(types.NewDocument("locale", "en", "numericOrdering", true)),
		},
		"Empty": {
			doc:    must.NotFail(types.NewDocument()),
			errMsg: "BSON field 'collation' is an empty object",
		},
		"MissingLocale": {
			doc:    must.NotFail(types.NewDocument("strength", int32(2))),
			errMsg: `Missing expected field "locale"`,
		},
		"InvalidLocale": {
			doc:    must.NotFail(types.NewDocument("locale", "invalid locale!")),
			errMsg: `Field 'locale' is invalid in: { locale: "invalid locale!" }`,
		},
		"WrongLocaleType": {
			doc:    must.NotFail(types.NewDocument("locale", int32(1))),
			errMsg: "BSON field 'collation.locale' is the wrong type 'int', expected type 'string'",
		},
		"StrengthTooLarge": {
			doc:    must.NotFail(types.NewDocument("locale", "en", "strength", int32(6))),
			errMsg: "BSON field 'collation.strength' value must be >= 1 and <= 5, actual value '6'",
		},
		"CaseLevelUnsupported": {
			doc:    must.NotFail(types.NewDocument("locale", "en", "caseLevel", true)),
			errMsg: "collation.caseLevel true is not supported yet",
		},
		"BackwardsUnsupported": {
			doc:    must.NotFail(types.NewDocument("locale", "en", "backwards", true)),
			errMsg: "collation.backwards true is not supported yet",
		},
		"UnknownField": {
			doc:    must.NotFail(types.NewDocument("locale", "en", "unknown", int32(1))),
			errMsg: "BSON field 'collation.unknown' is an unknown field",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			collation, err := NewCollation(tc.doc)

			if tc.errMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)

				return
			}

			require.NoError(t, err)

			if tc.nilCollation {
				assert.Nil(t, collation)
			} else {
				assert.NotNil(t, collation)
			}
		})
	}
}

func TestCollationCompare(t *testing.T) {
	t.Parallel()

	caseInsensitive := must.NotFail(NewCollation(must.NotFail(types.NewDocument(
		"locale", "en", "strength", int32(2),
	))))
	require.NotNil(t, caseInsensitive)

	assert.Zero(t, caseInsensitive.Compare("FerretDB", "ferretdb"))
	assert.Negative(t, caseInsensitive.Compare("apple", "Banana"))
	assert.Positive(t, caseInsensitive.Compare("Banana", "apple"))

	caseSensitive := must.NotFail(NewCollation(must.NotFail(types.NewDocument(
		"locale", "en",
	))))
	require.NotNil(t, caseSensitive)

	assert.NotZero(t, caseSensitive.Compare("FerretDB", "ferretdb"))

	numeric := must.NotFail(NewCollation(must.NotFail(types.NewDocument(
		"locale", "en", "numericOrdering", true,
	))))
	require.NotNil(t, numeric)

	assert.Negative(t, numeric.Compare("2", "10"))
}

func TestSortDocumentsWithCollation(t *testing.T) {
	t.Parallel()

	docs := []*types.Document{
		must.NotFail(types.NewDocument("_id", int32(1), "v", "banana")),
		must.NotFail(types.NewDocument("_id", int32(2), "v", "Apple")),
		must.NotFail(types.NewDocument("_id", int32(3), "v", "cherry")),
	}

	collation := must.NotFail(NewCollation(must.NotFail(types.NewDocument(
		"locale", "en", "strength", int32(2),
	))))

	err := SortDocumentsWithCollation(docs, must.NotFail(types.NewDocument("v", int32(1))), collation)
	require.NoError(t, err)

	var values []any
	for _, doc := range docs {
		values = append(values, must.NotFail(doc.Get("v")))
	}

	assert.Equal(t, []any{"Apple", "banana", "cherry"}, values)
}
