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

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/integration/setup"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuthentication(t *testing.T) {
	ctx, uri := setup.GetTargetURI(t)

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	require.NoError(t, err)

	err = client.Connect(ctx)
	require.NoError(t, err)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)
}
