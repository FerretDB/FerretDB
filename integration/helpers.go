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

// Package integration provides FerretDB integration tests.
package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

//go:generate ../bin/stringer  -type compatTestCaseResultType

// compatTestCaseResultType represents compatibility test case result type.
//
// It is used to avoid errors with invalid queries making tests pass.
type compatTestCaseResultType int

const (
	// Test case should return non-empty result at least for one collection/provider.
	nonEmptyResult compatTestCaseResultType = iota

	// Test case should return empty result for all collections/providers.
	emptyResult
)

// Convert converts given driver value (bson.D, bson.A, etc) to FerretDB types package value.
//
// It then can be used with all types helpers such as testutil.AssertEqual.
func Convert(t testing.TB, v any) any {
	t.Helper()

	switch v := v.(type) {
	// composite types
	case primitive.D:
		doc := must.NotFail(types.NewDocument())
		for _, e := range v {
			doc.Set(e.Key, Convert(t, e.Value))
		}
		return doc
	case primitive.A:
		arr := types.MakeArray(len(v))
		for _, e := range v {
			arr.Append(Convert(t, e))
		}
		return arr

	// scalar types (in the same order as in types package)
	case float64:
		return v
	case string:
		return v
	case primitive.Binary:
		return types.Binary{
			Subtype: types.BinarySubtype(v.Subtype),
			B:       v.Data,
		}
	case primitive.ObjectID:
		return types.ObjectID(v)
	case bool:
		return v
	case primitive.DateTime:
		return v.Time()
	case nil:
		return types.Null
	case primitive.Regex:
		return types.Regex{
			Pattern: v.Pattern,
			Options: v.Options,
		}
	case int32:
		return v
	case primitive.Timestamp:
		return types.NewTimestamp(time.Unix(int64(v.T), 0), uint32(v.I))
	case int64:
		return v
	default:
		t.Fatalf("unexpected type %T", v)
		panic("not reached")
	}
}

// ConvertDocument converts given driver's document to FerretDB's *types.Document.
func ConvertDocument(t testing.TB, doc bson.D) *types.Document {
	t.Helper()

	v := Convert(t, doc)

	var res *types.Document
	require.IsType(t, res, v)
	return v.(*types.Document)
}

// ConvertDocuments converts given driver's documents slice to FerretDB's []*types.Document.
func ConvertDocuments(t testing.TB, docs []bson.D) []*types.Document {
	t.Helper()

	res := make([]*types.Document, len(docs))
	for i, doc := range docs {
		res[i] = ConvertDocument(t, doc)
	}
	return res
}

// AssertEqualDocuments asserts that two documents are equal in a way that is useful for tests
// (NaNs are equal, etc).
//
// See testutil.AssertEqual for details.
func AssertEqualDocuments(t testing.TB, expected, actual bson.D) bool {
	t.Helper()

	expectedDoc := ConvertDocument(t, expected)
	actualDoc := ConvertDocument(t, actual)
	return testutil.AssertEqual(t, expectedDoc, actualDoc)
}

// AssertEqualDocumentsSlice asserts that two document slices are equal in a way that is useful for tests
// (NaNs are equal, etc).
//
// See testutil.AssertEqual for details.
func AssertEqualDocumentsSlice(t testing.TB, expected, actual []bson.D) bool {
	t.Helper()

	expectedDocs := ConvertDocuments(t, expected)
	actualDocs := ConvertDocuments(t, actual)
	return testutil.AssertEqualSlices(t, expectedDocs, actualDocs)
}

// AssertEqualError asserts that the expected error is the same as the actual (ignoring the Raw part).
func AssertEqualError(t testing.TB, expected mongo.CommandError, actual error) bool {
	t.Helper()

	a, ok := actual.(mongo.CommandError)
	if !ok {
		return assert.Equal(t, expected, actual)
	}

	// set expected fields that might be helpful in the test output
	require.Nil(t, expected.Raw)
	expected.Raw = a.Raw

	return assert.Equal(t, expected, a)
}

// AssertEqualAltError asserts that the expected error is the same as the actual (ignoring the Raw part);
// the alternative error message may be provided if FerretDB is unable to produce exactly the same text as MongoDB.
//
// In general, error messages should be the same. Exceptions include:
//
//   - MongoDB typos (e.g. "sortto" instead of "sort to");
//   - MongoDB values formatting (e.g. we don't want to write additional code to format
//     `{ $slice: { a: { b: 3 }, b: "string" } }` exactly the same way).
//
// In any case, the alternative error message returned by FerretDB should not mislead users.
func AssertEqualAltError(t testing.TB, expected mongo.CommandError, altMessage string, actual error) bool {
	t.Helper()

	a, ok := actual.(mongo.CommandError)
	if !ok {
		return assert.Equal(t, expected, actual)
	}

	// set expected fields that might be helpful in the test output
	require.Nil(t, expected.Raw)
	expected.Raw = a.Raw

	if assert.ObjectsAreEqual(expected, a) {
		return true
	}

	expected.Message = altMessage
	return assert.Equal(t, expected, a)
}

// AssertEqualWriteError asserts that the expected error is the same as the actual.
func AssertEqualWriteError(t *testing.T, expected mongo.WriteError, actual error) bool {
	t.Helper()

	we, ok := actual.(mongo.WriteException)
	if !ok {
		return assert.Equal(t, expected, actual)
	}

	if len(we.WriteErrors) != 1 {
		return assert.Equal(t, expected, actual)
	}

	a := we.WriteErrors[0]

	// set expected fields that might be helpful in the test output
	require.Nil(t, expected.Raw)
	expected.Raw = a.Raw

	return assert.Equal(t, expected, a)
}

// AssertEqualAltWriteError asserts that the expected error is the same as the actual;
// the alternative error message may be provided if FerretDB is unable to produce exactly the same text as MongoDB.
func AssertEqualAltWriteError(t *testing.T, expected mongo.WriteError, altMessage string, actual error) bool {
	t.Helper()

	we, ok := actual.(mongo.WriteException)
	if !ok {
		return assert.Equal(t, expected, actual)
	}

	if len(we.WriteErrors) != 1 {
		return assert.Equal(t, expected, actual)
	}

	a := we.WriteErrors[0]

	// set expected fields that might be helpful in the test output
	require.Nil(t, expected.Raw)
	expected.Raw = a.Raw

	if assert.ObjectsAreEqual(expected, a) {
		return true
	}

	expected.Message = altMessage
	return assert.Equal(t, expected, a)
}

// UnsetRaw returns error with all Raw fields unset. It returns nil if err is nil.
//
// Error is checked using a regular type assertion; wrapped errors (errors.As) are not checked.
func UnsetRaw(t testing.TB, err error) error {
	t.Helper()

	switch err := err.(type) {
	case mongo.CommandError:
		err.Raw = nil
		return err

	case mongo.WriteException:
		if err.WriteConcernError != nil {
			err.WriteConcernError.Raw = nil
		}
		for i, we := range err.WriteErrors {
			we.Raw = nil
			err.WriteErrors[i] = we
		}
		err.Raw = nil
		return err

	default:
		return err
	}
}

// CollectIDs returns all _id values from given documents.
//
// The order is preserved.
func CollectIDs(t testing.TB, docs []bson.D) []any {
	t.Helper()

	ids := make([]any, len(docs))
	for i, doc := range docs {
		id, ok := doc.Map()["_id"]
		require.True(t, ok)
		ids[i] = id
	}

	return ids
}

// CollectKeys returns document keys.
//
// The order is preserved.
func CollectKeys(t testing.TB, doc bson.D) []string {
	t.Helper()

	res := make([]string, len(doc))
	for i, e := range doc {
		res[i] = e.Key
	}

	return res
}

// FetchAll fetches all documents from the cursor, closing it.
func FetchAll(t testing.TB, ctx context.Context, cursor *mongo.Cursor) []bson.D {
	var res []bson.D
	err := cursor.All(ctx, &res)
	require.NoError(t, cursor.Close(ctx))
	require.NoError(t, err)
	return res
}

// FindAll returns all documents from the given collection sorted by _id.
func FindAll(t testing.TB, ctx context.Context, collection *mongo.Collection) []bson.D {
	opts := options.Find().SetSort(bson.D{{"_id", 1}})
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	require.NoError(t, err)

	return FetchAll(t, ctx, cursor)
}

// errorTextContains returns true if the error message contains at least one element of the given text slice.
// This function should be used to highlight the places where we do not have proper error checks yet
// but compare texts instead.
func errorTextContains(err error, texts ...string) bool {
	for _, text := range texts {
		if strings.Contains(err.Error(), text) {
			return true
		}
	}

	return false
}
