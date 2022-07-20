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

package tigris

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestSmoke(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.FixedScalars)

	var doc bson.D
	err := collection.FindOne(ctx, bson.D{{"_id", "fixed_double"}}).Decode(&doc)
	require.NoError(t, err)
	integration.AssertEqualDocuments(t, bson.D{{"_id", "fixed_double"}, {"double_value", 42.13}}, doc)
}
