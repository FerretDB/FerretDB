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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/testutil"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestDiffDatabaseName(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, dbName := range map[string]string{
		"ReservedPrefix": "_ferretdb_xxx",
		"NonLatin":       "データベース",
	} {
		name, dbName := name, dbName
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cName := testutil.CollectionName(t)

			err := collection.Database().Client().Database(dbName).CreateCollection(ctx, cName)

			if setup.IsMongoDB(t) {
				require.NoError(t, err)

				err = collection.Database().Client().Database(dbName).Drop(ctx)
				require.NoError(t, err)

				return
			}

			expected := mongo.CommandError{
				Name:    "InvalidNamespace",
				Code:    73,
				Message: fmt.Sprintf(`Invalid namespace specified '%s.%s'`, dbName, cName),
			}
			AssertEqualCommandError(t, expected, err)
		})
	}
}

func TestDiffCollectionName(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		collection  string
		excludeSpan bool
	}{
		"ReservedPrefix": {"_ferretdb_xxx", false},
		"NonUTF-8":       {string([]byte{0xff, 0xfe, 0xfd}), true}, // invalid UTF-8 won't be accepted by OTLP
	}

	t.Run("CreateCollection", func(t *testing.T) {
		for name, tc := range testcases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ctx, collection := setup.Setup(t)

				if tc.excludeSpan {
					var span trace.Span
					ctx, span = otel.Tracer("").Start(ctx, name+"ToExclude")
					span.SetAttributes(observability.ExclusionAttribute)

					defer span.End()
				}

				err := collection.Database().CreateCollection(ctx, tc.collection)

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				expected := mongo.CommandError{
					Name:    "InvalidNamespace",
					Code:    73,
					Message: fmt.Sprintf(`Invalid collection name: %s`, tc.collection),
				}
				AssertEqualCommandError(t, expected, err)
			})
		}
	})

	t.Run("RenameCollection", func(t *testing.T) {
		for name, tc := range testcases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ctx, collection := setup.Setup(t)

				if tc.excludeSpan {
					var span trace.Span
					ctx, span = otel.Tracer("").Start(ctx, name+"ToExclude")
					span.SetAttributes(observability.ExclusionAttribute)

					defer span.End()
				}

				fromName := testutil.CollectionName(t)
				err := collection.Database().CreateCollection(ctx, fromName)
				require.NoError(t, err)

				dbName := collection.Database().Name()
				command := bson.D{
					{"renameCollection", dbName + "." + fromName},
					{"to", dbName + "." + tc.collection},
				}

				err = collection.Database().Client().Database("admin").RunCommand(ctx, command).Err()

				if setup.IsMongoDB(t) {
					require.NoError(t, err)
					return
				}

				expected := mongo.CommandError{
					Name:    "IllegalOperation",
					Code:    20,
					Message: fmt.Sprintf(`error with target namespace: Invalid collection name: %s`, tc.collection),
				}
				AssertEqualCommandError(t, expected, err)
			})
		}
	})
}
