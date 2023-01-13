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

package types

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// TestDocumentValidateData covers ValidateData method.
// Proper testing of validation requires integration tests,
// see https://github.com/FerretDB/dance/tree/main/tests/diff for more examples.
func TestDocumentValidateData(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		doc    *Document
		reason error
	}{
		"Valid": {
			doc:    must.NotFail(NewDocument("_id", "1", "foo", "bar")),
			reason: nil,
		},
		"KeyIsNotUTF8": {
			doc:    must.NotFail(NewDocument("\xf4\x90\x80\x80", "bar")), //  the key is out of range for UTF-8
			reason: errors.New(`invalid key: "\xf4\x90\x80\x80" (not a valid UTF-8 string)`),
		},
		"KeyContainsDollarSign": {
			doc:    must.NotFail(NewDocument("$v", "bar")),
			reason: errors.New(`invalid key: "$v" (key must not contain '$' sign)`),
		},
		"KeyContainsDotSign": {
			doc:    must.NotFail(NewDocument("v.foo", "bar")),
			reason: errors.New(`invalid key: "v.foo" (key must not contain '.' sign)`),
		},
		"DuplicateKeys": {
			doc:    must.NotFail(NewDocument("_id", "1", "foo", "bar", "foo", "baz")),
			reason: errors.New(`invalid key: "foo" (duplicate keys are not allowed)`),
		},
		"Inf+": {
			doc:    must.NotFail(NewDocument("v", math.Inf(1))),
			reason: errors.New(`invalid value: { "v": +Inf } (infinity values are not allowed)`),
		},
		"Inf-": {
			doc:    must.NotFail(NewDocument("v", math.Inf(-1))),
			reason: errors.New(`invalid value: { "v": -Inf } (infinity values are not allowed)`),
		},
		"NoID": {
			doc:    must.NotFail(NewDocument("foo", "bar")),
			reason: errors.New(`invalid document: document must contain '_id' field`),
		},
		"Array": {
			doc:    must.NotFail(NewDocument("_id", &Array{[]any{"foo", "bar"}})),
			reason: errors.New("The '_id' value cannot be of type array"),
		},
		"Regex": {
			doc:    must.NotFail(NewDocument("_id", Regex{Pattern: "regex$"})),
			reason: errors.New("The '_id' value cannot be of type regex"),
		},
		"NestedArray": {
			doc: must.NotFail(NewDocument(
				"_id", "1",
				"foo", must.NotFail(NewArray("bar", must.NotFail(NewArray("baz")))),
			)),
			reason: errors.New(`invalid value: { "foo": [ "bar", [ "baz" ] ] } (nested arrays are not supported)`),
		},
		"NestedDocumentNestedArray": {
			doc: must.NotFail(NewDocument(
				"_id", "1",
				"foo", must.NotFail(NewDocument(
					"bar", must.NotFail(NewArray("baz", must.NotFail(NewArray("qaz")))),
				)),
			)),
			reason: errors.New(`invalid value: { "bar": [ "baz", [ "qaz" ] ] } (nested arrays are not supported)`),
		},
		"ArrayDocumentNestedArray": {
			doc: must.NotFail(NewDocument(
				"_id", "1",
				"foo", must.NotFail(NewArray(
					must.NotFail(NewDocument(
						"bar", must.NotFail(NewArray("baz", must.NotFail(NewArray("qaz")))),
					)),
				)),
			)),
			reason: errors.New(`invalid value: { "bar": [ "baz", [ "qaz" ] ] } (nested arrays are not supported)`),
		},
		"NegativeZeroFromUpdate": {
			doc: must.NotFail(NewDocument("v", math.Copysign(0, -1))),
			reason: errors.New(`update produces invalid value: { "v": -0 } ` +
				`(-0 is not supported)`),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.doc.ValidateData()
			if tc.reason == nil {
				assert.NoError(t, err)
			} else {
				var ve *ValidationError
				require.True(t, errors.As(err, &ve))
				assert.Equal(t, tc.reason, ve.reason)
			}
		})
	}
}
