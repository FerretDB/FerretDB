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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
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

// convert converts given driver value (bson.D, bson.A, etc) to FerretDB types package value.
//
// It then can be used with all types helpers such as testutil.AssertEqual.
func convert(t testtb.TB, v any) any {
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
func ConvertDocument(t testtb.TB, doc bson.D) *types.Document {
	t.Helper()

	v := convert(t, doc)

	var res *types.Document
	require.IsType(t, res, v)
	return v.(*types.Document)
}

// ConvertDocuments converts given driver's documents slice to FerretDB's []*types.Document.
func ConvertDocuments(t testtb.TB, docs []bson.D) []*types.Document {
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
func AssertEqualDocuments(t testtb.TB, expected, actual bson.D) bool {
	t.Helper()

	expectedDoc := ConvertDocument(t, expected)
	actualDoc := ConvertDocument(t, actual)
	return testutil.AssertEqual(t, expectedDoc, actualDoc)
}

// AssertEqualDocumentsSlice asserts that two document slices are equal in a way that is useful for tests
// (NaNs are equal, etc).
//
// See testutil.AssertEqual for details.
func AssertEqualDocumentsSlice(t testtb.TB, expected, actual []bson.D) bool {
	t.Helper()

	expectedDocs := ConvertDocuments(t, expected)
	actualDocs := ConvertDocuments(t, actual)
	return testutil.AssertEqualSlices(t, expectedDocs, actualDocs)
}

// AssertEqualCommandError asserts that the expected error is the same as the actual (ignoring the Raw part).
func AssertEqualCommandError(t testtb.TB, expected mongo.CommandError, actual error) bool {
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

// AssertEqualWriteError asserts that actual is a WriteException containing exactly one expected error (ignoring the Raw part).
func AssertEqualWriteError(t testtb.TB, expected mongo.WriteError, actual error) bool {
	t.Helper()

	we, ok := actual.(mongo.WriteException) //nolint:errorlint // do not inspect error chain
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
func AssertMatchesCommandError(t testtb.TB, expected, actual error) {
	t.Helper()

	a, ok := actual.(mongo.CommandError) //nolint:errorlint // do not inspect error chain
	require.Truef(t, ok, "actual is %T, not mongo.CommandError", actual)

	e, ok := expected.(mongo.CommandError) //nolint:errorlint // do not inspect error chain
	require.Truef(t, ok, "expected is %T, not mongo.CommandError", expected)

	a.Raw = nil
	e.Raw = nil

	actualMessage := a.Message
	a.Message = e.Message

	if !AssertEqualCommandError(t, e, a) {
		t.Logf("actual message: %s", actualMessage)
	}
}

// AssertMatchesWriteError asserts that both errors are WriteExceptions containing exactly one WriteError,
// and those WriteErrors are equal, except messages (and ignoring the Raw part).
func AssertMatchesWriteError(t testtb.TB, expected, actual error) {
	t.Helper()

	a, ok := actual.(mongo.WriteException) //nolint:errorlint // do not inspect error chain
	require.Truef(t, ok, "actual is %T, not mongo.WriteException", actual)
	require.Lenf(t, a.WriteErrors, 1, "actual is %v, expected one mongo.WriteError", a.WriteErrors)

	e, ok := expected.(mongo.WriteException) //nolint:errorlint // do not inspect error chain
	require.Truef(t, ok, "expected is %T, not mongo.WriteException", expected)
	require.Lenf(t, e.WriteErrors, 1, "expected is %v, expected one mongo.WriteError", e.WriteErrors)

	aErr := a.WriteErrors[0]
	eErr := e.WriteErrors[0]

	aErr.Raw = nil
	eErr.Raw = nil

	actualMessage := aErr.Message
	aErr.Message = eErr.Message

	if !AssertEqualWriteError(t, eErr, aErr) {
		t.Logf("actual message: %s", actualMessage)
	}
}

// AssertMatchesBulkException asserts that both errors are BulkWriteExceptions containing the same number of WriteErrors,
// and those WriteErrors are equal, except messages (and ignoring the Raw part).
func AssertMatchesBulkException(t testtb.TB, expected, actual error) {
	t.Helper()

	a, ok := actual.(mongo.BulkWriteException) //nolint:errorlint // do not inspect error chain
	require.Truef(t, ok, "actual is %T, not mongo.BulkWriteException", actual)

	e, ok := expected.(mongo.BulkWriteException) //nolint:errorlint // do not inspect error chain
	require.Truef(t, ok, "expected is %T, not mongo.BulkWriteException", expected)

	for i, we := range a.WriteErrors {
		expectedWe := e.WriteErrors[i]

		expectedWe.Message = we.Message
		expectedWe.Raw = we.Raw

		assert.Equal(t, expectedWe, we)
	}
}

// AssertEqualAltCommandError asserts that the expected MongoDB error is the same as the actual (ignoring the Raw part);
// the alternative error message may be provided if FerretDB is unable to produce exactly the same text as MongoDB.
//
// In general, error messages should be the same. Exceptions include:
//
//   - MongoDB typos (e.g. "sortto" instead of "sort to");
//   - MongoDB values formatting (e.g. we don't want to write additional code to format
//     `{ $slice: { a: { b: 3 }, b: "string" } }` exactly the same way).
//
// In any case, the alternative error message returned by FerretDB should not mislead users.
func AssertEqualAltCommandError(t testtb.TB, expected mongo.CommandError, altMessage string, actual error) bool {
	t.Helper()

	a, ok := actual.(mongo.CommandError)
	if !ok {
		return assert.Equal(t, expected, actual)
	}

	// set expected fields that might be helpful in the test output
	require.Nil(t, expected.Raw)
	expected.Raw = a.Raw

	if setup.IsMongoDB(t) || altMessage == "" {
		return assert.Equal(t, expected, a)
	}

	if assert.ObjectsAreEqual(expected, a) {
		return true
	}

	expected.Message = altMessage
	return assert.Equal(t, expected, a)
}

// AssertEqualAltWriteError asserts that the expected MongoDB error is the same as the actual;
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

	if setup.IsMongoDB(t) || altMessage == "" {
		return assert.Equal(t, expected, a)
	}

	if assert.ObjectsAreEqual(expected, a) {
		return true
	}

	expected.Message = altMessage
	return assert.Equal(t, expected, a)
}

// UnsetRaw returns error with all Raw fields unset. It returns nil if err is nil.
//
// Error is checked using a regular type assertion; wrapped errors (errors.As) are not checked.
func UnsetRaw(t testtb.TB, err error) error {
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
func CollectIDs(t testtb.TB, docs []bson.D) []any {
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
func CollectKeys(t testtb.TB, doc bson.D) []string {
	t.Helper()

	res := make([]string, len(doc))
	for i, e := range doc {
		res[i] = e.Key
	}

	return res
}

// FetchAll fetches all documents from the cursor, closing it.
func FetchAll(t testtb.TB, ctx context.Context, cursor *mongo.Cursor) []bson.D {
	var res []bson.D
	err := cursor.All(ctx, &res)
	require.NoError(t, cursor.Close(ctx))
	require.NoError(t, err)
	return res
}

// FindAll returns all documents from the given collection sorted by _id.
func FindAll(t testtb.TB, ctx context.Context, collection *mongo.Collection) []bson.D {
	opts := options.Find().SetSort(bson.D{{"_id", 1}})
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	require.NoError(t, err)

	return FetchAll(t, ctx, cursor)
}

// generateDocuments generates documents with _id ranging from startID to endID.
// It returns bson.A and []bson.D both containing same bson.D documents.
func generateDocuments(startID, endID int32) (bson.A, []bson.D) {
	var arr bson.A
	var docs []bson.D

	for i := startID; i < endID; i++ {
		arr = append(arr, bson.D{{"_id", i}})
		docs = append(docs, bson.D{{"_id", i}})
	}

	return arr, docs
}

// CreateNestedDocument creates a mock BSON document that consists of nested arrays and documents.
// The nesting level is based on integer parameter.
func CreateNestedDocument(n int) bson.D {
	return createNestedDocument(n, false).(bson.D)
}

// createNestedDocument creates the nested n times object that consists of
// documents and arrays. If the arr is true, the root value will be array.
//
// This function should be used only internally.
// To generate values for tests please use
// exported CreateNestedDocument function.
func createNestedDocument(n int, arr bool) any {
	var child any

	if n > 0 {
		child = createNestedDocument(n-1, !arr)
	}

	if arr {
		return bson.A{child}
	}

	return bson.D{{"v", child}}
}
