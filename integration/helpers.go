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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

//go:generate ../bin/stringer  -type compatTestCaseResultType

// documentValidationFailureCode is returned by Tigris schema validation code.
// TODO tigris provider should only use collections that does not produce
// validation error for each test case.
// https://github.com/FerretDB/FerretDB/issues/2253
const documentValidationFailureCode = 121

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

// convert converts given driver value (bson.D, bson.A, etc) to FerretDB types package value.
//
// It then can be used with all types helpers such as testutil.AssertEqual.
func convert(t testing.TB, v any) any {
	t.Helper()

	switch v := v.(type) {
	// composite types
	case primitive.D:
		doc := types.MakeDocument(len(v))
		for _, e := range v {
			doc.Set(e.Key, convert(t, e.Value))
		}
		return doc
	case primitive.A:
		arr := types.MakeArray(len(v))
		for _, e := range v {
			arr.Append(convert(t, e))
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

	v := convert(t, doc)

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

// AssertEqualError is a deprecated alias for AssertEqualCommandError.
//
// Deprecated: use AssertEqualCommandError instead.
func AssertEqualError(t testing.TB, expected mongo.CommandError, actual error) bool {
	return AssertEqualCommandError(t, expected, actual)
}

// AssertEqualError asserts that the expected error is the same as the actual (ignoring the Raw part).
func AssertEqualCommandError(t testing.TB, expected mongo.CommandError, actual error) bool {
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

// AssertEqualError asserts that actual is a WriteException containing exactly one expected error (ignoring the Raw part).
func AssertEqualWriteError(t testing.TB, expected mongo.WriteError, actual error) bool {
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

// AssertMatchesCommandError asserts that both errors are equal CommandErrors,
// except messages (and ignoring the Raw part).
func AssertMatchesCommandError(t testing.TB, expected, actual error) {
	t.Helper()

	var a mongo.CommandError
	require.ErrorAs(t, actual, &a)

	var e mongo.CommandError
	require.ErrorAs(t, expected, &e)

	a.Raw = nil
	e.Raw = nil

	actualMessage := a.Message
	a.Message = e.Message
	if !AssertEqualError(t, e, a) {
		t.Logf("actual message: %s", actualMessage)
	}
}

// AssertMatchesWriteError asserts error codes are the same.
//
// TODO check not only code; make is look similar to AssertMatchesCommandError above.
// https://github.com/FerretDB/FerretDB/issues/2545
func AssertMatchesWriteError(t testing.TB, expected, actual error) {
	t.Helper()

	var aErr, eErr mongo.WriteException

	if ok := errors.As(actual, &aErr); !ok || len(aErr.WriteErrors) != 1 {
		assert.Equal(t, expected, actual)
		return
	}

	if ok := errors.As(expected, &eErr); !ok || len(eErr.WriteErrors) != 1 {
		assert.Equal(t, expected, actual)
		return
	}

	assert.Equal(t, eErr.WriteErrors[0].Code, aErr.WriteErrors[0].Code)
}

// AssertEqualAltError is a deprecated alias for AssertEqualAltCommandError.
//
// Deprecated: use AssertEqualAltCommandError instead.
func AssertEqualAltError(t testing.TB, expected mongo.CommandError, altMessage string, actual error) bool {
	return AssertEqualAltCommandError(t, expected, altMessage, actual)
}

// AssertEqualAltCommandError asserts that the expected error is the same as the actual (ignoring the Raw part);
// the alternative error message may be provided if FerretDB is unable to produce exactly the same text as MongoDB.
//
// In general, error messages should be the same. Exceptions include:
//
//   - MongoDB typos (e.g. "sortto" instead of "sort to");
//   - MongoDB values formatting (e.g. we don't want to write additional code to format
//     `{ $slice: { a: { b: 3 }, b: "string" } }` exactly the same way).
//
// In any case, the alternative error message returned by FerretDB should not mislead users.
func AssertEqualAltCommandError(t testing.TB, expected mongo.CommandError, altMessage string, actual error) bool {
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

	case mongo.BulkWriteException:
		if err.WriteConcernError != nil {
			err.WriteConcernError.Raw = nil
		}

		for i, we := range err.WriteErrors {
			we.Raw = nil
			err.WriteErrors[i] = we
		}

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
