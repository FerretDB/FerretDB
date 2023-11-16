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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestDiffErrorMessages(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		DatabaseName: "admin",
	})
	ctx, db := s.Ctx, s.Collection.Database()

	err := db.RunCommand(ctx, bson.D{{"getParameter", bson.D{{"allParameters", "1"}}}}).Err()

	if setup.IsMongoDB(t) {
		expected := mongo.CommandError{
			Code: 14,
			Name: "TypeMismatch",
			Message: "BSON field 'getParameter.allParameters' is the wrong type 'string', " +
				"expected types '[bool, long, int, decimal, double']",
		}
		AssertEqualCommandError(t, expected, err)

		return
	}

	expected := mongo.CommandError{
		Code:    14,
		Name:    "TypeMismatch",
		Message: "BSON field 'allParameters' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
	}
	AssertEqualCommandError(t, expected, err)
}
