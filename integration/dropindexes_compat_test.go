package integration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestIndexesDrop(t *testing.T) {
	t.Helper()

	setup.SkipForTigrisWithReason(t, "Indexes drop is not supported for Tigris")

	for name, tc := range map[string]struct {
		models        []mongo.IndexModel          // creates index if not nil
		dropIndexName string                      // name of a single index to drop
		dropAll       bool                        // set true for drop all indexes, if true dropIndexName must be empty.
		opts          *options.DropIndexesOptions // required
		altErrorMsg   string                      // optional, alternative error message in case of error
		resultType    compatTestCaseResultType    // defaults to nonEmptyResult
		skip          string                      // optional, skip test with a specified reason
	}{
		"DropAll": {
			dropAll: true,
		},
		"ID": {
			dropIndexName: "_id_",
			resultType:    emptyResult,
		},
		"ValueAscending": {
			dropIndexName: "v_1",
		},
		"Value": {
			dropIndexName: "v_-1",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			if tc.dropAll {
				require.Empty(t, tc.dropIndexName, "index name must be empty when dropping all indexes")
			}

			t.Helper()
			t.Parallel()

			// Use single provider for drop indexes test.
			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.Composites},
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					if tc.models != nil {
						_, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.models)
						_, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.models)
						require.NoError(t, compatErr)
						require.NoError(t, targetErr)
					}

					var targetRes, compatRes bson.Raw
					var targetErr, compatErr error

					if tc.dropAll {
						targetRes, targetErr = targetCollection.Indexes().DropAll(ctx, tc.opts)
						compatRes, compatErr = compatCollection.Indexes().DropAll(ctx, tc.opts)
					} else {
						targetRes, targetErr = targetCollection.Indexes().DropOne(ctx, tc.dropIndexName, tc.opts)
						compatRes, compatErr = compatCollection.Indexes().DropOne(ctx, tc.dropIndexName, tc.opts)
					}

					if tc.altErrorMsg != "" {
						AssertMatchesCommandError(t, compatErr, targetErr)

						var expectedErr mongo.CommandError
						require.True(t, errors.As(compatErr, &expectedErr))
						expectedErr.Raw = nil
						AssertEqualAltError(t, expectedErr, tc.altErrorMsg, targetErr)
					} else {
						require.Equal(t, compatErr, targetErr)
					}

					assert.Equal(t, compatRes, targetRes)

					if compatErr == nil {
						nonEmptyResults = true
					}

					// List indexes to see they are identical after creation.
					targetCur, targetErr := targetCollection.Indexes().List(ctx)
					compatCur, compatErr := compatCollection.Indexes().List(ctx)

					require.NoError(t, compatErr)
					assert.Equal(t, compatErr, targetErr)

					targetIndexes := FetchAll(t, ctx, targetCur)
					compatIndexes := FetchAll(t, ctx, compatCur)

					assert.Equal(t, compatIndexes, targetIndexes)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestIndexesDropRunCommand(t *testing.T) {
	t.Helper()

	setup.SkipForTigrisWithReason(t, "Indexes creation is not supported for Tigris")

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)
	targetCollection := targetCollections[0]
	compatCollection := compatCollections[0]

	for name, tc := range map[string]struct { //nolint:vet // for readability
		collectionName string
		index          any
		resultType     compatTestCaseResultType // defaults to nonEmptyResult
		skip           string                   // optional, skip test with a specified reason
		altErrorMsg    string
	}{
		"invalid-index-type": {
			collectionName: targetCollection.Name(),
			index:          true,
			resultType:     emptyResult,
		},
		"invalid-index": {
			collectionName: targetCollection.Name(),
			index:          "non-existent",
			resultType:     emptyResult,
		},
		"invalid-collection": {
			collectionName: "non-existent",
			resultType:     emptyResult,
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Helper()
			t.Parallel()

			command := bson.D{
				{"dropIndexes", tc.collectionName},
				{"index", tc.index},
			}

			var targetRes bson.D
			targetErr := targetCollection.Database().RunCommand(ctx, command).Decode(&targetRes)

			var compatRes bson.D
			compatErr := compatCollection.Database().RunCommand(ctx, command).Decode(&compatRes)

			if tc.altErrorMsg != "" {
				AssertMatchesCommandError(t, compatErr, targetErr)

				var expectedErr mongo.CommandError
				require.True(t, errors.As(compatErr, &expectedErr))
				expectedErr.Raw = nil
				AssertEqualAltError(t, expectedErr, tc.altErrorMsg, targetErr)
			} else {
				require.Equal(t, compatErr, targetErr)
			}

			if tc.resultType == emptyResult {
				require.Nil(t, targetRes)
				require.Nil(t, compatRes)
			}

			assert.Equal(t, compatRes, targetRes)

			targetErr = targetCollection.Database().RunCommand(
				ctx, bson.D{{"listIndexes", tc.collectionName}},
			).Decode(&targetRes)

			compatErr = compatCollection.Database().RunCommand(
				ctx, bson.D{{"listIndexes", tc.collectionName}},
			).Decode(&compatRes)

			assert.Equal(t, compatRes, targetRes)

			AssertMatchesCommandError(t, compatErr, targetErr)
		})
	}
}
