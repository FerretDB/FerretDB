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
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// databaseName returns valid database name for given test.
func databaseName(t *testing.T) string {
	t.Helper()

	name := strings.ToLower(t.Name())
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")

	require.Less(t, len(name), 64)
	return name
}

// collectionName returns valid collection name for given test.
func collectionName(t *testing.T) string {
	t.Helper()

	name := strings.ToLower(t.Name())
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, " ", "-")

	require.Less(t, len(name), 64)
	return name
}

func ScalarsData() map[string]any {
	return map[string]any{
		"double":                   42.13,
		"double-zero":              0.0,
		"double-negative-zero":     math.Copysign(0, -1),
		"double-max":               math.MaxFloat64,
		"double-smallest":          math.SmallestNonzeroFloat64,
		"double-positive-infinity": math.Inf(+1),
		"double-negative-infinity": math.Inf(-1),
		"double-nan":               math.NaN(),

		"string":       "foo",
		"string-empty": "",

		// no Document
		// no Array

		"binary":       primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}},
		"binary-empty": primitive.Binary{},

		// no Undefined

		"bool-false": false,
		"bool-true":  true,

		"datetime":          time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC),
		"datetime-epoch":    time.Unix(0, 0),
		"datetime-year-min": time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
		"datetime-year-max": time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC),

		"null": nil,

		"regex":       primitive.Regex{Pattern: "foo", Options: "i"},
		"regex-empty": primitive.Regex{},

		// no DBPointer
		// no JavaScript code
		// no Symbol
		// no JavaScript code w/ scope

		"int32":      int32(42),
		"int32-zero": int32(0),
		"int32-max":  int32(math.MaxInt32),
		"int32-min":  int32(math.MinInt32),

		"timestamp":   primitive.Timestamp{T: 42, I: 13},
		"timestamp-i": primitive.Timestamp{I: 1},

		"int64":      int64(42),
		"int64-zero": int64(0),
		"int64-max":  int64(math.MaxInt64),
		"int64-min":  int64(math.MinInt64),

		// no 128-bit decimal floating point (yet)

		// no Min key
		// no Max key
	}
}

func Scalars(ctx context.Context, t *testing.T, db *mongo.Database) {
	collection := db.Collection(collectionName(t))
	for id, v := range ScalarsData() {
		_, err := collection.InsertOne(ctx, bson.D{{"_id", id}, {"value", v}})
		require.NoError(t, err)
	}
}
