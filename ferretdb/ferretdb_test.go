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

package ferretdb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ExampleNew is a testable example for Run func.
func ExampleNew() {
	ctx, cancel := context.WithCancel(context.Background())
	conf := Config{Handler: "pg", PostgresURL: "postgres://postgres@127.0.0.1:5432/ferretdb"}

	fdb := New(conf)
	go func() {
		err := fdb.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	uri := fdb.GetConnectionString()

	var err error
	var client *mongo.Client
	if client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri)); err != nil {
		panic(err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		panic(err)
	}
	client.Disconnect(ctx)
	cancel()
	fmt.Println(uri)
	// Output: mongodb://127.0.0.1:27017
}
