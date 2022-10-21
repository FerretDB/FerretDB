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
)

func Example_tcp() {
	f, err := New(&Config{
		ListenAddr:    "127.0.0.1:17027",
		Handler:       "pg",
		PostgreSQLURL: "postgres://postgres@127.0.0.1:5432/ferretdb",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		log.Print(f.Run(ctx))
		close(done)
	}()

	uri := f.MongoDBURI()
	fmt.Println(uri)

	// Use MongoDB URI as usual. For example:
	//
	// import "go.mongodb.org/mongo-driver/mongo"
	//
	// [...]
	//
	// mongo.Connect(ctx, options.Client().ApplyURI(uri)

	cancel()
	<-done

	// Output: mongodb://127.0.0.1:17027/
}

func Example_unix() {
	f, err := New(&Config{
		ListenUnix:    "/tmp/ferretdb-27017.sock",
		Handler:       "pg",
		PostgreSQLURL: "postgres://postgres@127.0.0.1:5432/ferretdb",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})

	go func() {
		log.Print(f.Run(ctx))
		close(done)
	}()

	uri := f.MongoDBURI()
	fmt.Println(uri)

	// Use MongoDB URI as usual.

	cancel()
	<-done

	// Output: mongodb://%2Ftmp%2Fferretdb-27017.sock
}
