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

package cursor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// getComparableCursorResponse takes the response from the query that generates the cursors,
// returns the cursor ID and the response without cursor.id field.
// If cursor.id field doesn't exist, it returns the response as is and nil cursor ID.
func getComparableCursorResponse(t testing.TB, res bson.D) (bson.D, any) {
	t.Helper()

	var cursorID any
	var comparableRes bson.D

	for _, field := range res {
		switch field.Key {
		case "cursor":
			var ok bool
			cursor, ok := field.Value.(bson.D)
			require.True(t, ok)

			var cursorWithoutID bson.D

			for _, cursorField := range cursor {
				switch cursorField.Key {
				case "id":
					cursorID = cursorField.Value
				default:
					cursorWithoutID = append(cursorWithoutID, cursorField)
				}
			}

			comparableRes = append(comparableRes, bson.E{Key: "cursor", Value: cursorWithoutID})
		default:
			comparableRes = append(comparableRes, field)
		}
	}

	return comparableRes, cursorID
}
