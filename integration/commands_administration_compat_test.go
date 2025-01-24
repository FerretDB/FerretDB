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
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestListCollectionsCompat(t *testing.T) {
	t.Parallel()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)

	filterNames := make(bson.A, len(targetCollections))
	for i, n := range targetCollections {
		filterNames[i] = n.Name()
	}

	// We should remove shuffle there once it is implemented in the setup.
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/825

	rand.Shuffle(len(filterNames), func(i, j int) { filterNames[i], filterNames[j] = filterNames[j], filterNames[i] })
	filterNames = filterNames[:len(filterNames)-1]
	require.NotEmpty(t, filterNames)

	filter := bson.D{{
		"name", bson.D{{
			"$in", filterNames,
		}},
	}}

	compat, err := compatCollections[0].Database().ListCollections(ctx, filter)
	require.NoError(t, err)
	defer compat.Close(ctx)

	var compatRes []bson.D
	err = compat.All(ctx, &compatRes)
	require.NoError(t, err)

	compatNames := make([]string, len(compatRes))
	for i, doc := range compatRes {
		compatNames[i] = doc.Map()["name"].(string)
	}

	require.True(t, slices.IsSorted(compatNames), "compat collections are not sorted")

	target, err := targetCollections[0].Database().ListCollections(ctx, filter)
	require.NoError(t, err)
	defer target.Close(ctx)

	var targetRes []bson.D
	err = target.All(ctx, &targetRes)
	require.NoError(t, err)

	assert.Equal(t, target.RemainingBatchLength(), compat.RemainingBatchLength())

	comparable := func(res []bson.D) []bson.D {
		var resComparable []bson.D

		for _, doc := range res {
			var docComparable bson.D

			for _, field := range doc {
				switch field.Key {
				case "info":
					info, ok := field.Value.(bson.D)
					require.True(t, ok)

					var infoComparable bson.D

					for _, infoField := range info {
						switch infoField.Key {
						case "uuid":
							uuid, uuidOk := infoField.Value.(primitive.Binary)
							require.True(t, uuidOk)
							assert.Equal(t, bson.TypeBinaryUUID, uuid.Subtype)
							assert.Len(t, uuid.Data, 16)
							infoComparable = append(infoComparable, bson.E{Key: infoField.Key, Value: primitive.Binary{}})
						default:
							infoComparable = append(infoComparable, infoField)
						}
					}

					docComparable = append(docComparable, bson.E{Key: field.Key, Value: infoComparable})

				default:
					docComparable = append(docComparable, field)
				}
			}

			resComparable = append(resComparable, docComparable)
		}

		return resComparable
	}

	AssertEqualDocumentsSlice(t, comparable(compatRes), comparable(targetRes))
}
