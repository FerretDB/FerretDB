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
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/util/testutil"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestFindAndModifyCompatSimple(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"EmptyQueryRemove": {
			command: bson.D{
				{"query", bson.D{}},
				{"remove", true},
			},
		},
		"NewDoubleNonZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "double-smallest"}}},
				{"update", bson.D{{"_id", "double-smallest"}, {"v", float64(43)}}},
				{"new", float64(42)},
			},
		},
		"NewDoubleZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "double-zero"}}},
				{"update", bson.D{{"_id", "double-zero"}, {"v", 43.0}}},
				{"new", float64(0)},
			},
		},
		"NewIntNonZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
				{"new", int32(11)},
			},
		},
		"NewIntZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32-zero"}}},
				{"update", bson.D{{"_id", "int32-zero"}, {"v", int32(43)}}},
				{"new", int32(0)},
			},
		},
		"NewLongNonZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"_id", "int64"}, {"v", int64(43)}}},
				{"new", int64(11)},
			},
		},
		"NewLongZero": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64-zero"}}},
				{"update", bson.D{{"_id", "int64-zero"}, {"v", int64(43)}}},
				{"new", int64(0)},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatErrors(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"NotEnoughParameters": {
			command:    bson.D{},
			resultType: emptyResult,
		},
		"UpdateAndRemove": {
			command: bson.D{
				{"update", bson.D{}},
				{"remove", true},
			},
			resultType: emptyResult,
		},
		"NewAndRemove": {
			command: bson.D{
				{"new", true},
				{"remove", true},
			},
			resultType: emptyResult,
		},
		"InvalidUpdateType": {
			command: bson.D{
				{"query", bson.D{}},
				{"update", "123"},
			},
			resultType: emptyResult,
		},
		"InvalidMaxTimeMSType": {
			command: bson.D{
				{"maxTimeMS", "string"},
			},
			resultType: emptyResult,
		},
		"DuplicateID": {
			command: bson.D{
				{"query", bson.D{{"non-existent", "val"}}},
				{"update", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
				{"upsert", true},
			},
			providers:  []shareddata.Provider{shareddata.Int32s},
			resultType: emptyResult,
		},
		"InvalidID": {
			command: bson.D{
				{"query", bson.D{{"non-existent", "val"}}},
				{"update", bson.D{{"_id", primitive.Regex{Pattern: "[a-z]*[0-9]"}}, {"v", int32(43)}}},
				{"upsert", true},
			},
			providers:  []shareddata.Provider{shareddata.Int32s},
			resultType: emptyResult,
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUpdate(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"Replace": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"_id", "int64"}, {"v", int64(43)}}},
			},
		},
		"ReplaceWithoutID": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"v", int64(43)}}},
			},
		},
		"ReplaceReturnNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
				{"new", true},
			},
		},
		"NotExistedIdInQuery": {
			command: bson.D{
				{"query", bson.D{{"_id", "no-such-id"}}},
				{"update", bson.D{{"v", int32(43)}}},
			},
		},
		"NotExistedIdNotInQuery": {
			command: bson.D{
				{"query", bson.D{{"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", 0}}}},
					bson.D{{"v", bson.D{{"$lt", 0}}}},
				}}}},
				{"update", bson.D{{"v", int32(43)}}},
			},
		},
		"UpdateOperatorSet": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"$set", bson.D{{"v", int64(43)}}}}},
			},
		},
		"UpdateOperatorSetReturnNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"$set", bson.D{{"v", int64(43)}}}}},
				{"new", true},
			},
		},
		"EmptyUpdate": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"v", bson.D{}}}},
			},
		},
		"Conflict": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"update", bson.D{{"$invalid", "non-existent-field"}}},
			},
			resultType: emptyResult,
		},
		"OperatorConflict": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"update", bson.D{
					{"$set", bson.D{{"v", 4}}},
					{"$inc", bson.D{{"v", 4}}},
				}},
			},
			resultType: emptyResult,
		},
		"NoConflict": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{
					{"$set", bson.D{{"v", 4}}},
					{"$inc", bson.D{{"foo", 4}}},
				}},
			},
		},
		"EmptyKey": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{
					{"$set", bson.D{{"", 4}}},
					{"$inc", bson.D{{"", 4}}},
				}},
			},
			resultType: emptyResult,
		},
		"EmptyKeyAndKey": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{
					{"$set", bson.D{{"", 4}}},
					{"$inc", bson.D{{"v", 4}}},
				}},
			},
			resultType: emptyResult,
		},
		"InvalidOperator": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"update", bson.D{{"$invalid", "non-existent-field"}}},
			},
			resultType: emptyResult,
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatDotNotation(t *testing.T) {
	testCases := map[string]findAndModifyCompatTestCase{
		"Conflict": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-two-fields"}}},
				{"update", bson.D{
					{"$set", bson.D{{"v.0.field", 4}}},
					{"$inc", bson.D{{"v.0.field", 4}}},
				}},
			},
			resultType: emptyResult,
		},
		"NoConflict": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-two-fields"}}},
				{"update", bson.D{
					{"$set", bson.D{{"v.0.field", 4}}},
					{"$inc", bson.D{{"v.0.foo", 4}}},
				}},
			},
		},
		"NoIndex": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-two-fields"}}},
				{"update", bson.D{
					{"$set", bson.D{{"v.0.field", 4}}},
					{"$inc", bson.D{{"v.field", 4}}},
				}},
			},
		},
		"ParentConflict": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-two-fields"}}},
				{"update", bson.D{
					{"$set", bson.D{{"v.0.field", 4}}},
					{"$inc", bson.D{{"v", 4}}},
				}},
			},
			resultType: emptyResult,
		},

		"ConflictKey": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"update", bson.D{
					{"$set", bson.D{{"v", "val"}}},
					{"$min", bson.D{{"v.foo", "val"}}},
				}},
			},
			resultType: emptyResult,
		},
		"ConflictKeyPrefix": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"update", bson.D{
					{"$set", bson.D{{"v.foo", "val"}}},
					{"$min", bson.D{{"v", "val"}}},
				}},
			},
			resultType: emptyResult,
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUpdateSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"NonExistentExistsTrue": {
			command: bson.D{
				{"query", bson.D{{"non-existent", bson.D{{"$exists", true}}}}},
				{"update", bson.D{{"$set", bson.D{{"v", "foo"}}}}},
			},
		},
		"NonExistentExistsFalse": {
			command: bson.D{
				{"query", bson.D{{"non-existent", bson.D{{"$exists", false}}}}},
				{"update", bson.D{{"$set", bson.D{{"v", "foo"}}}}},
			},
		},
		"ExistsTrue": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", true}}}}},
				{"update", bson.D{{"$set", bson.D{{"v", "foo"}}}}},
			},
		},
		"ExistsFalse": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"update", bson.D{{"$set", bson.D{{"v", "foo"}}}}},
			},
		},
		"UpdateIDNoQuery": {
			command: bson.D{
				{"update", bson.D{{"$set", bson.D{{"_id", "int32"}}}}},
			},
		},
		"UpdateExistingID": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"$set", bson.D{{"_id", "int32-1"}}}}},
			},
		},
		"UpdateSameID": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"$set", bson.D{{"_id", "int32"}}}}},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUnset(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"NonExistentExistsT": {
			command: bson.D{
				{"query", bson.D{{"non-existent", bson.D{{"$exists", true}}}}},
				{"update", bson.D{{"$unset", bson.D{{"v", ""}}}}},
			},
		},
		"NonExistentExistsF": {
			command: bson.D{
				{"query", bson.D{{"non-existent", bson.D{{"$exists", false}}}}},
				{"update", bson.D{{"$unset", bson.D{{"v", ""}}}}},
			},
		},
		"ExistsTrue": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", true}}}}},
				{"update", bson.D{{"$unset", bson.D{{"v", ""}}}}},
			},
		},
		"ExistsFalse": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"update", bson.D{{"$unset", bson.D{{"v", ""}}}}},
			},
		},
		"UnsetNonExistentField": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"$unset", bson.D{{"non-existent-field", ""}}}}},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUpdateCurrentDate(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"NotDocument": {
			command: bson.D{
				{"query", bson.D{{"_id", "datetime"}}},
				{"update", bson.D{{"$currentDate", 1}}},
			},
			providers:  []shareddata.Provider{shareddata.DateTimes},
			resultType: emptyResult,
		},
		"UnknownOption": {
			command: bson.D{
				{"query", bson.D{{"_id", "datetime"}}},
				{"update", bson.D{{"$currentDate", bson.D{{"v", bson.D{{"foo", int32(1)}}}}}}},
			},
			providers:  []shareddata.Provider{shareddata.DateTimes},
			resultType: emptyResult,
		},
		"NoType": {
			command: bson.D{
				{"query", bson.D{{"_id", "datetime"}}},
				{"update", bson.D{{"$currentDate", bson.D{{"v", bson.D{}}}}}},
			},
			providers:  []shareddata.Provider{shareddata.DateTimes},
			resultType: emptyResult,
		},
		"InvalidType": {
			command: bson.D{
				{"query", bson.D{{"_id", "datetime"}}},
				{"update", bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", int32(1)}}}}}}},
			},
			providers:  []shareddata.Provider{shareddata.DateTimes},
			resultType: emptyResult,
		},
		"UnknownType": {
			command: bson.D{
				{"query", bson.D{{"_id", "datetime"}}},
				{"update", bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "unknown"}}}}}}},
			},
			providers:  []shareddata.Provider{shareddata.DateTimes},
			resultType: emptyResult,
		},
		"InvalidValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "datetime"}}},
				{"update", bson.D{{"$currentDate", bson.D{{"v", 1}}}}},
			},
			providers:  []shareddata.Provider{shareddata.DateTimes},
			resultType: emptyResult,
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUpdateRename(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"NotDocument": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"$rename", 1}}},
			},
			resultType: emptyResult,
		},
		"NonStringTargetField": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"$rename", bson.D{{"v", 0}}}}},
			},
			resultType: emptyResult,
		},
		"SameTargetField": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"$rename", bson.D{{"v", "v"}}}}},
			},
			resultType: emptyResult,
		},
		"DuplicateSource": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"$rename", bson.D{{"v", "w"}, {"v", "x"}}}}},
			},
			resultType: emptyResult,
		},
		"DuplicateTarget": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64"}}},
				{"update", bson.D{{"$rename", bson.D{{"v", "w"}, {"x", "w"}}}}},
			},
			resultType: emptyResult,
		},
	}

	testFindAndModifyCompat(t, testCases)
}

// TestFindAndModifyCompatSort tests how various sort orders are handled.
//
// Add more tests for sort.
// TODO https://github.com/FerretDB/FerretDB/issues/2168
func TestFindAndModifyCompatSort(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"DotNotation": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$in", bson.A{"array-documents-nested", "array-documents-nested-duplicate"}}}}}},
				{"update", bson.D{{"$set", bson.D{{"v.0.foo.0.bar", "baz"}}}}},
				{"sort", bson.D{{"v.0.foo", 1}, {"_id", 1}}},
			},
		},
		"DotNotationIndex": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$in", bson.A{"array-documents-nested", "array-documents-nested-duplicate"}}}}}},
				{"update", bson.D{{"$set", bson.D{{"v.0.foo.0.bar", "baz"}}}}},
				{"sort", bson.D{{"v.0.foo.0.bar", 1}, {"_id", 1}}},
			},
		},
		"DotNotationNonExistent": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$in", bson.A{"array-documents-nested", "array-documents-nested-duplicate"}}}}}},
				{"update", bson.D{{"$set", bson.D{{"v.0.foo.0.bar", "baz"}}}}},
				{"sort", bson.D{{"invalid.foo", 1}, {"_id", 1}}},
			},
		},
		"DotNotationMissingField": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$in", bson.A{"array-documents-nested", "array-documents-nested-duplicate"}}}}}},
				{"update", bson.D{{"$set", bson.D{{"v.0.foo.0.bar", "baz"}}}}},
				{"sort", bson.D{{"v..foo", 1}, {"_id", 1}}},
			},
			resultType: emptyResult,
		},
		"DollarPrefixedFieldName": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$in", bson.A{"array-documents-nested", "array-documents-nested-duplicate"}}}}}},
				{"update", bson.D{{"$set", bson.D{{"v.0.foo.0.bar", "baz"}}}}},
				{"sort", bson.D{{"$v.foo", 1}, {"_id", 1}}},
			},
			resultType: emptyResult,
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUpsert(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"UpsertNoSuchDocument": {
			command: bson.D{
				{"query", bson.D{{"_id", "no-such-doc"}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
				{"new", true},
			},
		},
		"UpsertNoReplaceDocument": {
			command: bson.D{
				{"query", bson.D{{"_id", "no-such-doc"}}},
				{"update", bson.D{{"v", 43.13}}},
				{"upsert", true},
				{"new", true},
			},
		},
		"UpsertReplace": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"v", 43.13}}},
				{"upsert", true},
			},
		},
		"UpsertReplaceReturnNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"v", 43.13}}},
				{"upsert", true},
				{"new", true},
			},
		},
		"ExistsNew": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"upsert", true},
				{"update", bson.D{{"_id", "replaced"}, {"v", "replaced"}}},
				{"new", true},
			},
		},
		"ExistsFalse": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"upsert", true},
				{"update", bson.D{{"_id", "replaced"}, {"v", "replaced"}}},
			},
		},
		"UpdateID": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", true}}}}},
				{"upsert", true},
				{"update", bson.D{{"_id", "int32"}, {"v", "replaced"}}},
			},
		},
		"UpdateDifferentID": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", true}}}}},
				{"upsert", true},
				{"update", bson.D{{"_id", "replaced"}, {"v", "replaced"}}},
			},
			resultType: emptyResult, // _id must be an immutable field
		},
		"ExistsTrue": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", true}}}}},
				{"upsert", true},
				{"update", bson.D{{"v", "replaced"}}},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUpsertSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"Upsert": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
			},
		},
		"UpsertNew": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
				{"new", true},
			},
		},
		"UpsertNonExistent": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"v", "43"}}}}},
			},
		},
		"UpsertNewNonExistent": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"v", "43"}}}}},
				{"new", true},
			},
		},
		"NonExistentExistsFalse": {
			command: bson.D{
				{"query", bson.D{{"non-existent", bson.D{{"$exists", false}}}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"v", "foo"}}}}},
			},
		},
		"ExistsTrue": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", true}}}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"v", "foo"}}}}},
			},
		},
		"UpsertID": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"_id", "double"}}}}},
			},
			resultType: emptyResult, // _id must be an immutable field
		},
		"UpsertIDNoQuery": {
			command: bson.D{
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"_id", "int32"}, {"v", int32(2)}}}}},
			},
		},
		"UpsertExistingID": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"_id", "int32-1"}, {"v", int32(2)}}}}},
			},
			resultType: emptyResult,
		},
		"UpsertSameID": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"_id", "int32"}, {"v", int32(2)}}}}},
			},
		},
		"UpsertQueryOperatorEq": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$eq", "non-existent"}}}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"new", "val"}}}}},
			},
		},
		"UpsertQueryOperatorMixed": {
			command: bson.D{
				{"query", bson.D{
					{"_id", bson.D{{"$eq", "non-existent"}}},
					{"v", bson.D{{"$lt", 43}}},
					{"non_existent", int32(0)},
				}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"new", "val"}}}}},
			},
		},
		"UpsertQueryObject": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}, {"v", bson.D{{"k1", "v1"}}}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"new", "val"}}}}},
			},
		},
		"UpsertQueryObjectNested": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}, {"v", bson.D{{"k1", "v1"}, {"k2", bson.D{{"k21", "v21"}}}}}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"new", "val"}}}}},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatSetOnInsert(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"IDExists": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"upsert", true},
				{"new", true},
				{"update", bson.D{{"$setOnInsert", bson.D{{"new", "val"}}}}},
			},
			providers: []shareddata.Provider{shareddata.Int32s},
		},
		"IDNotExists": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"upsert", true},
				{"new", true},
				{"update", bson.D{{"$setOnInsert", bson.D{{"new", "val"}}}}},
			},
		},
		"UpsertFalse": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"upsert", false},
				{"new", true},
				{"update", bson.D{{"$setOnInsert", bson.D{{"new", "val"}}}}},
			},
		},
		"SetWithSetOnInsert": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"upsert", true},
				{"new", true},
				{"update", bson.D{
					{"$set", bson.D{{"new", "val"}}},
					{"$setOnInsert", bson.D{{"v", int32(42)}}},
				}},
			},
		},
		"ApplySetSkipSetOnInsert": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"upsert", true},
				{"new", true},
				{"update", bson.D{
					{"$set", bson.D{{"new", "val"}}},
					{"$setOnInsert", bson.D{{"v", int32(43)}}},
				}},
			},
			providers: []shareddata.Provider{shareddata.Int32s},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatUpsertUnset(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"NonExistentExistsT": {
			command: bson.D{
				{"query", bson.D{{"non-existent", bson.D{{"$exists", true}}}}},
				{"upsert", true},
				{"update", bson.D{
					{"$unset", bson.D{{"v", ""}}},
					{"$set", bson.D{{"_id", "upserted"}}}, // to have the same _id for target and compat
				}},
			},
		},
		"NonExistentExistsF": {
			command: bson.D{
				{"query", bson.D{{"non-existent", bson.D{{"$exists", false}}}}},
				{"upsert", true},
				{"update", bson.D{{"$unset", bson.D{{"v", ""}}}}},
			},
		},
		"ExistsTrue": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", true}}}}},
				{"upsert", true},
				{"update", bson.D{{"$unset", bson.D{{"v", ""}}}}},
			},
		},
		"ExistsFalse": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"upsert", true},
				{"update", bson.D{
					{"$unset", bson.D{{"v", ""}}},
					{"$set", bson.D{{"_id", "upserted"}}}, // to have the same _id for target and compat
				}},
			},
		},
		"UnsetNonExistentField": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"upsert", true},
				{"update", bson.D{{"$unset", bson.D{{"non-existent-field", ""}}}}},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatRemove(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"Remove": {
			command: bson.D{
				{"query", bson.D{{"_id", "double"}}},
				{"remove", true},
			},
		},
		"RemoveEmptyQueryResult": {
			command: bson.D{
				{
					"query",
					bson.D{{
						"$and",
						bson.A{
							bson.D{{"v", bson.D{{"$gt", 0}}}},
							bson.D{{"v", bson.D{{"$lt", 0}}}},
						},
					}},
				},
				{"remove", true},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyCompatReplacementDoc(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"Basic": {
			command: bson.D{
				{"update", bson.D{{"v", int32(43)}}},
			},
		},
		"EmptyDoc": {
			command: bson.D{
				{"update", bson.D{}},
			},
		},
		"FilterAndUpsertTrue": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"update", bson.D{{"v", int32(43)}}},
				{"upsert", true},
			},
		},
		"WithUpdateOp": {
			command: bson.D{
				{"update", bson.D{{"v", int32(43)}, {"$set", bson.D{{"test", int32(0)}}}}},
			},
			resultType: emptyResult,
		},
		"SameId": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"_id", "int32"}, {"v", int32(43)}}},
			},
		},
		"DifferentId": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"_id", "non-existent"}, {"v", int32(43)}}},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

func TestFindAndModifyFieldOrder(t *testing.T) {
	t.Parallel()

	testCases := map[string]findAndModifyCompatTestCase{
		"MultipleOps": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"update", bson.D{
					{"$set", bson.D{{"b", 2}, {"d", 4}}},
					{"$setOnInsert", bson.D{{"c", "3"}}},
					{"$inc", bson.D{{"a", 1}}},
					{"$min", bson.D{{"f", 5}}},
					{"$max", bson.D{{"h", 7}}},
					{"$mul", bson.D{{"g", 6}}},
					{"$bit", bson.D{{"m", bson.D{{"xor", 1}}}}},
					{"$push", bson.D{{"i", 8}}},
					{"$addToSet", bson.D{{"j", 9}}},
				}},
				{"upsert", true},
				{"new", true},
			},
		},
		"UpsertFalse": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{
					{"$set", bson.D{{"b", 2}, {"d", 4}, {"v", int32(43)}}},
					{"$inc", bson.D{{"a", 1}}},
					{"$push", bson.D{{"e", 5}}},
					{"$setOnInsert", bson.D{{"c", "3"}}},
				}},
				{"upsert", false},
			},
			providers: []shareddata.Provider{shareddata.Int32s},
		},
		"NestedField": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"update", bson.D{
					{"$set", bson.D{{"x", bson.D{{"c", 3}, {"b", 2}}}}},
					{"$setOnInsert", bson.D{{"a", "1"}}},
				}},
				{"upsert", true},
				{"new", true},
			},
		},
		"NestedFieldDotNotation": {
			command: bson.D{
				{"query", bson.D{{"_id", "non-existent"}}},
				{"update", bson.D{
					{"$set", bson.D{{"a.y", 2}}},
					{"$inc", bson.D{{"a.x", 1}}},
					{"$setOnInsert", bson.D{{"a.z", "3"}}},
				}},
				{"upsert", true},
				{"new", true},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

// findAndModifyCompatTestCase describes findAndModify compatibility test case.
type findAndModifyCompatTestCase struct {
	command    bson.D
	resultType compatTestCaseResultType // defaults to nonEmptyResult
	providers  []shareddata.Provider    // defaults to shareddata.AllProviders()

	skip string // skips test if non-empty
}

// testFindAndModifyCompat tests findAndModify compatibility test cases.
func testFindAndModifyCompat(t *testing.T, testCases map[string]findAndModifyCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			providers := shareddata.AllProviders()
			if tc.providers != nil {
				providers = tc.providers
			}

			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{Providers: providers})

			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					targetCommand := bson.D{{"findAndModify", targetCollection.Name()}}
					targetCommand = append(targetCommand, tc.command...)
					if targetCommand.Map()["sort"] == nil {
						targetCommand = append(targetCommand, bson.D{{"sort", bson.D{{"_id", 1}}}}...)
					}

					compatCommand := bson.D{{"findAndModify", compatCollection.Name()}}
					compatCommand = append(compatCommand, tc.command...)
					if compatCommand.Map()["sort"] == nil {
						compatCommand = append(compatCommand, bson.D{{"sort", bson.D{{"_id", 1}}}}...)
					}

					var targetMod, compatMod bson.D
					var targetErr, compatErr error
					targetErr = targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetMod)
					compatErr = compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatMod)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						AssertMatchesError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					compat := ConvertDocument(t, compatMod)
					compat.Remove("$clusterTime")
					compat.Remove("operationTime")

					target := ConvertDocument(t, targetMod)
					testutil.AssertEqual(t, compat, target)

					// To make sure that the results of modification are equal,
					// find all the documents in target and compat collections and compare that they are the same
					opts := options.Find().SetSort(bson.D{{"_id", 1}})
					targetCursor, targetErr := targetCollection.Find(ctx, bson.D{}, opts)
					compatCursor, compatErr := compatCollection.Find(ctx, bson.D{}, opts)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						targetErr = UnsetRaw(t, targetErr)
						compatErr = UnsetRaw(t, compatErr)
						assert.Equal(t, compatErr, targetErr)
						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					targetRes := FetchAll(t, ctx, targetCursor)
					compatRes := FetchAll(t, ctx, compatCursor)

					t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatRes))
					t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetRes))
					AssertEqualDocumentsSlice(t, compatRes, targetRes)

					if len(targetRes) > 0 || len(compatRes) > 0 {
						nonEmptyResults = true
					}
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}
