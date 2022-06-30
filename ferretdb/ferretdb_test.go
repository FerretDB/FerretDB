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
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Example() {
	f, err := New(&Config{
		Handler:       "pg",
		PostgreSQLURL: "postgres://postgres@127.0.0.1:5432/ferretdb",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	go f.Run(ctx)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(f.MongoDBURI()))
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.D{{
		"name",
		bson.D{{
			"$not",
			bson.D{{
				"$regex",
				primitive.Regex{Pattern: "test.*"},
			}},
		}},
	}}
	collections, err := client.ListDatabaseNames(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(collections)
	// Output: [admin public]
}
