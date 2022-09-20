package integration

import (
	"testing"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type user struct {
	ID   primitive.ObjectID `bson:"_id"`
	Name string
	Age  int8
}

// deleteCompatTestCase describes delete compatibility test case.
type insertCompatTestCase struct {
	documents  []any
	ordered    bool // defaults to true
	noOfResult int
	// resultType compatTestCaseResultType // defaults to nonEmptyResult
	skip string // skips test if non-empty
}

func TestInsertCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]insertCompatTestCase{
		"Empty": {
			documents:  []any{},
			noOfResult: 0,
			ordered:    true,
		},

		"One": {
			documents:  []any{user{ID: primitive.NewObjectID(), Name: "John", Age: 10}},
			noOfResult: 1,
			ordered:    true,
		},

		"Many": {
			documents: []any{
				user{ID: primitive.NewObjectID(), Name: "John", Age: 10},
				user{ID: primitive.NewObjectID(), Name: "Alex", Age: 28},
				user{ID: primitive.NewObjectID(), Name: "Elena", Age: 23},
			},
			noOfResult: 3,
			ordered:    true,
		},

		// "UnorderedMany": {
		// 	documents: []any{
		// 		user{ID: must.NotFail(primitive.ObjectIDFromHex(`6329888190ff4c1bf93d72db`)), Name: "John", Age: 10},
		// 		user{ID: must.NotFail(primitive.ObjectIDFromHex(`6329888190ff4c1bf93d72db`)), Name: "Alex", Age: 28},
		// 		user{ID: must.NotFail(primitive.ObjectIDFromHex(`6329888190ff4c1bf93d72db`)), Name: "Elena", Age: 23},
		// 	},
		// 	ordered:    false,
		// 	noOfResult: 0,
		// },

		// "Two":    {},
		// "TwoAll": {},
		// "TwoAllOrdered": {
		// 	ordered: true,
		// },

		// "OrderedMany": {
		// 	documents: []any{
		// 		user{ID: primitive.NewObjectID(), Name: "John", Age: 10},
		// 		user{ID: must.NotFail(primitive.ObjectIDFromHex(`6327b8c1fed4f5470dbb775c`)), Name: "Alex", Age: 28},
		// 		user{ID: must.NotFail(primitive.ObjectIDFromHex(`6327b8c1fed4f5470dbb775c`)), Name: "Elena", Age: 23},
		// 	},
		// 	ordered:    true,
		// 	noOfResult: 2,
		// },
		// "UnorderedError": {},

		// "OrderedTwoErrors": {
		// 	ordered: true,
		// },
		// "UnorderedTwoErrors": {},

		// "OrderedAllErrors": {
		// 	ordered:    true,
		// 	resultType: emptyResult,
		// },
		// "UnorderedAllErrors": {
		// 	resultType: emptyResult,
		// },
	}

	testInsertCompat(t, testCases)
}

// testInsertCompat tests insert compatibility test cases.
func testInsertCompat(t *testing.T, testCases map[string]insertCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			// Use per-test setup because deletes modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			// var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					opts := options.InsertMany().SetOrdered(tc.ordered)

					targetRes, targetErr := targetCollection.InsertMany(ctx, tc.documents, opts)
					compatRes, compatErr := compatCollection.InsertMany(ctx, tc.documents, opts)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						targetErr = UnsetRaw(t, targetErr)
						compatErr = UnsetRaw(t, compatErr)
						assert.Equal(t, compatErr, targetErr)
					} else {
						assert.NotNil(t, targetRes)
						assert.NotNil(t, compatRes)
						assert.NoError(t, compatErr, "compat error")

						t.Logf("Target -> %v", targetRes.InsertedIDs)
						t.Logf("Compat -> %v", compatRes.InsertedIDs)

						assert.True(t,
							len(compatRes.InsertedIDs) == len(targetRes.InsertedIDs) &&
								len(compatRes.InsertedIDs) == tc.noOfResult)

						// if len(compatRes.InsertedIDs) > 0 && len(targetRes.InsertedIDs) > 0 {
						// 	nonEmptyResults = true
						// }
					}

					// assert.Equal(t, compatRes, targetRes)

					// targetDocs := FindAll(t, ctx, targetCollection)
					// compatDocs := FindAll(t, ctx, compatCollection)

					// t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatDocs))
					// t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetDocs))
					// AssertEqualDocumentsSlice(t, compatDocs, targetDocs)
				})
				// 		}

				// switch tc.resultType {
				// case nonEmptyResult:
				// 	assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be deleted)")
				// case emptyResult:
				// 	assert.False(t, nonEmptyResults, "expected empty results (no documents should be deleted)")
				// default:
				// 	t.Fatalf("unknown result type %v", tc.resultType)
				// }
			}
		})
	}
}
